// scheduler 独立调度进程：内置轻量调度引擎（中小企业无需部署 DolphinScheduler）。
//
// 职责：扫中台主库三张任务表 → 候选任务注册进调度器 → 到期时经 etcd 发现各 rpc 执行：
//   - dc_quality_job     （cron 非空）              → quality.rpc RunJob
//   - dc_ops_archive_task（cron 非空）              → ops.rpc   RunArchive
//   - dc_theme_mv        （refresh_mode=schedule）  → theme.rpc RefreshMaterializedView
//
// last_run_at / last_status 回写由各 rpc 内部完成（RunJob/RunArchive/RefreshMv 均已实现），
// 调度进程只负责"到点触发"。
//
// 注意：单实例运行（暂无分布式锁，勿多副本部署，后续接入）。
// 每轮巡检会重扫表并增量注册新任务（同名任务忽略后到者），删任务需重启生效。
package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"datacenter/internal/scheduler"
	"datacenter/ops/rpc/ops"
	"datacenter/ops/rpc/opsClient"
	"datacenter/quality/rpc/quality"
	"datacenter/quality/rpc/qualityClient"
	"datacenter/theme/rpc/theme"
	"datacenter/theme/rpc/themeClient"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

// envEtcd 各 rpc 服务发现用的 etcd 地址（逗号分隔）
const envEtcd = "SCHED_ETCD_HOSTS"

// defaultEtcd 默认 etcd 地址（与各 rpc 服务 etc 配置一致）
const defaultEtcd = "127.0.0.1:2379"

// defaultMvCron dc_theme_mv 表无 cron 列，schedule 模式统一默认每天 02:00 刷新
const defaultMvCron = "0 2 * * *"

// runTimeout 单次任务执行的兜底超时（rpc 同步执行，防止长任务挂死占住并发位）
const runTimeout = 10 * time.Minute

// rpcClients 三个目标服务的客户端（etcd 发现，NonBlock 启动不阻塞、调用时才建连）
type rpcClients struct {
	Quality qualityClient.Quality
	Ops     opsClient.Ops
	Theme   themeClient.Theme
}

func main() {
	// 优雅退出：SIGINT/SIGTERM → ctx 取消 → 调度器等在跑任务收尾后退出
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := scheduler.Open()
	if err != nil {
		logx.Errorf("[scheduler] 打开扫表连接失败（SCHED_MYSQL_DSN=%s）: %v", scheduler.DSN(), err)
		return
	}
	defer db.Close()

	cli := newClients()

	// 启动即扫表注册一轮，随后每 tick 重扫增量注册（新建任务无需重启调度进程）
	sch := scheduler.New(0)
	registerFromScan(ctx, sch, db, cli)
	go rescanLoop(ctx, sch, db, cli)

	sch.Run(ctx)
}

// newClients 经 etcd 构造 quality/ops/theme 三个 rpc 客户端
func newClients() rpcClients {
	hosts := strings.Split(envOr(envEtcd, defaultEtcd), ",")
	mk := func(key string) zrpc.Client {
		return zrpc.MustNewClient(zrpc.RpcClientConf{
			Etcd:     discov.EtcdConf{Hosts: hosts, Key: key},
			NonBlock: true, // 服务暂不可用不阻塞启动，调用时才报错
		})
	}
	return rpcClients{
		Quality: qualityClient.NewQuality(mk("quality.rpc")),
		Ops:     opsClient.NewOps(mk("ops.rpc")),
		Theme:   themeClient.NewTheme(mk("theme.rpc")),
	}
}

// registerFromScan 扫表并把候选任务注册进调度器（同名任务被忽略，天然幂等）。
// 按来源分派到对应 rpc 的执行函数；cron 为空（theme 场景）补默认 cron；
// 单条 cron 非法由 Scheduler.Register 跳过并记日志，不拖垮其余任务。
func registerFromScan(ctx context.Context, sch *scheduler.Scheduler, db *sql.DB, cli rpcClients) {
	rows, err := scheduler.ScanDueCandidates(ctx, db)
	if err != nil {
		logx.Errorf("[scheduler] 扫表部分失败（已注册成功部分的行）: %v", err)
	}
	var added int
	for _, row := range rows {
		task, err := buildTask(cli, row)
		if err != nil {
			logx.Errorf("[scheduler] 候选任务 %s:%d 构建失败，跳过: %v", row.Source, row.Id, err)
			continue
		}
		ok, err := sch.Register(task)
		if err != nil {
			continue // Register 内部已 Errorf（cron 非法等）
		}
		if ok {
			added++
			logx.Infof("[scheduler] 注册任务 %s（name=%s，beid=%d）", task.Name, row.Name, row.Beid)
		}
	}
	logx.Infof("[scheduler] 本轮扫表：候选 %d 条，新注册 %d 个，调度器在册 %d 个",
		len(rows), added, sch.Len())
}

// buildTask 把一行候选任务转成调度器 Task：按来源分派 rpc 执行函数
// （执行函数以触发时的 ctx 运行，收到调度器退出信号会随 ctx 取消而中止）
func buildTask(cli rpcClients, row scheduler.DueRow) (scheduler.Task, error) {
	task := scheduler.Task{
		Name:    row.Source + ":" + itoa(row.Id),
		Cron:    row.Cron,
		LastRun: row.LastRun,
	}
	if row.Cron == "" { // dc_theme_mv 无 cron 列，补默认（每天 02:00）
		task.Cron = defaultMvCron
	}
	switch row.Source {
	case "quality":
		task.Fn = func(ctx context.Context) error {
			c, cancel := context.WithTimeout(ctx, runTimeout)
			defer cancel()
			reply, err := cli.Quality.RunJob(c, &quality.IdReq{Id: row.Id, Beid: row.Beid})
			if err != nil {
				return err
			}
			logx.Infof("[scheduler] 质量任务 %d（beid=%d）执行完成：规则 %d 条，通过 %d，状态 %s",
				row.Id, row.Beid, reply.TotalRules, reply.Passed, reply.LastStatus)
			return nil
		}
	case "ops":
		task.Fn = func(ctx context.Context) error {
			c, cancel := context.WithTimeout(ctx, runTimeout)
			defer cancel()
			reply, err := cli.Ops.RunArchive(c, &ops.IdReq{Id: row.Id, Beid: row.Beid})
			if err != nil {
				return err
			}
			logx.Infof("[scheduler] 归档任务 %d（beid=%d）执行完成：归档 %d 行，删除 %d 行，状态 %s",
				row.Id, row.Beid, reply.ArchivedRows, reply.DeletedRows, reply.Status)
			return nil
		}
	case "theme":
		task.Fn = func(ctx context.Context) error {
			c, cancel := context.WithTimeout(ctx, runTimeout)
			defer cancel()
			reply, err := cli.Theme.RefreshMaterializedView(c, &theme.IdReq{Id: row.Id, Beid: row.Beid})
			if err != nil {
				return err
			}
			// 刷新失败以 ok=false 回告（rpc 内部已回写 last_status=failed），此处仅记日志不算调度失败
			logx.Infof("[scheduler] 物化视图 %d（beid=%d）刷新：ok=%v msg=%s 耗时=%dms",
				row.Id, row.Beid, reply.Ok, reply.Msg, reply.CostMs)
			return nil
		}
	default:
		return scheduler.Task{}, errUnknownSource
	}
	return task, nil
}

// rescanLoop 周期重扫表：增量注册新建的任务（同名忽略），与调度器同节奏巡检
func rescanLoop(ctx context.Context, sch *scheduler.Scheduler, db *sql.DB, cli rpcClients) {
	ticker := time.NewTicker(scheduler.DefaultTick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			registerFromScan(ctx, sch, db, cli)
		}
	}
}

// errUnknownSource 未知的任务来源（扫表 SQL 与分派逻辑不同步时的防御）
var errUnknownSource = errors.New("未知的任务来源")

// itoa 任务 id 转字符串（任务命名用）
func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}

// envOr 读环境变量，未配置返回默认值
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
