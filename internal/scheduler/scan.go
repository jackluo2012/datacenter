// scan.go 调度进程的扫表实现：直连中台主库（标准库 database/sql + go-sql-driver），
// 捞出三类"可定时执行"的候选任务行；是否真正到期由 scheduler 侧用 cron+last_run 判断，
// 本文件只负责查行（SQL 保持简单，不加专用 model —— 跨服务直连仅此一处）。
//
// 覆盖范围（与设计契约一致）：
//   - dc_quality_job     cron 非空                  → quality.rpc RunJob
//   - dc_ops_archive_task cron 非空                 → ops.rpc   RunArchive
//   - dc_theme_mv        refresh_mode='schedule'    → theme.rpc RefreshMaterializedView
//     （该表无 cron 列，cron 由调用方填默认值）
//   - dc_tag_group       圈群暂无 cron 字段，跳过（后续再接）
package scheduler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	// mysql 驱动注册到标准库 database/sql（go.mod 已有依赖，无需新增）
	_ "github.com/go-sql-driver/mysql"
)

// DefaultDSN 默认中台主库连接串（与 quality.rpc 等服务 etc 配置同库）
const DefaultDSN = "root:admin@tcp(127.0.0.1:3306)/datacenter?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"

// envDSN 扫表连接串的环境变量名
const envDSN = "SCHED_MYSQL_DSN"

// DSN 扫表连接串：优先环境变量 SCHED_MYSQL_DSN，未配置用默认值
func DSN() string {
	if v := os.Getenv(envDSN); v != "" {
		return v
	}
	return DefaultDSN
}

// Open 打开扫表连接（调用方负责 Close）
func Open() (*sql.DB, error) {
	return sql.Open("mysql", DSN())
}

// DueRow 扫表得到的一条候选到期任务
type DueRow struct {
	Source  string     // 来源：quality / ops / theme（用于任务命名与分派 rpc）
	Id      int64      // 业务表主键
	Beid    int64      // 租户 id（各 rpc 的 RunXxx 均要求 beid > 0）
	Name    string     // 任务名（日志展示）
	Cron    string     // cron 表达式（theme 表无此列，调用方补默认值）
	LastRun *time.Time // 上次执行时间（表列为 NULL 时为 nil → 启动即到期一次）
}

// ScanDueCandidates 扫三张任务表，返回候选任务行。
// 单表扫描失败不阻断其余表：返回已捞到的行 + 聚合错误（调用方记日志即可，
// 下一轮巡检会重扫，天然自愈临时性 DB 故障）。
func ScanDueCandidates(ctx context.Context, db *sql.DB) ([]DueRow, error) {
	rows, errs := make([]DueRow, 0), []error{}

	q, err := scanWithCron(ctx, db,
		`SELECT id, name, beid, cron, last_run_at
		   FROM dc_quality_job
		  WHERE cron <> ''`, "quality")
	if err != nil {
		errs = append(errs, fmt.Errorf("扫 dc_quality_job: %w", err))
	}
	rows = append(rows, q...)

	a, err := scanWithCron(ctx, db,
		`SELECT id, name, beid, cron, last_run_at
		   FROM dc_ops_archive_task
		  WHERE cron <> ''`, "ops")
	if err != nil {
		errs = append(errs, fmt.Errorf("扫 dc_ops_archive_task: %w", err))
	}
	rows = append(rows, a...)

	// 物化视图：计划刷新模式（表无 cron/name 之外的调度列，Cron 留空由调用方补默认值）
	m, err := scanThemeMvs(ctx, db)
	if err != nil {
		errs = append(errs, fmt.Errorf("扫 dc_theme_mv: %w", err))
	}
	rows = append(rows, m...)

	return rows, errors.Join(errs...)
}

// scanWithCron 扫"带 cron 列"的任务表（质量任务 / 归档任务共用，列序一致）
func scanWithCron(ctx context.Context, db *sql.DB, query, source string) ([]DueRow, error) {
	rs, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := make([]DueRow, 0, 16)
	for rs.Next() {
		var (
			id, beid int64
			name     string
			cron     sql.NullString
			lastRun  sql.NullTime
		)
		if err := rs.Scan(&id, &name, &beid, &cron, &lastRun); err != nil {
			return nil, err
		}
		row := DueRow{Source: source, Id: id, Beid: beid, Name: name, Cron: cron.String}
		if lastRun.Valid {
			lr := lastRun.Time
			row.LastRun = &lr
		}
		out = append(out, row)
	}
	return out, rs.Err()
}

// scanThemeMvs 扫物化视图：refresh_mode='schedule' 的行（无 cron 列，Cron 留空）
func scanThemeMvs(ctx context.Context, db *sql.DB) ([]DueRow, error) {
	query := `SELECT id, name, beid, last_refresh_at
		        FROM dc_theme_mv
		       WHERE refresh_mode = 'schedule'`
	rs, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := make([]DueRow, 0, 16)
	for rs.Next() {
		var (
			id, beid int64
			name     string
			lastRun  sql.NullTime
		)
		if err := rs.Scan(&id, &name, &beid, &lastRun); err != nil {
			return nil, err
		}
		row := DueRow{Source: "theme", Id: id, Beid: beid, Name: name}
		if lastRun.Valid {
			lr := lastRun.Time
			row.LastRun = &lr
		}
		out = append(out, row)
	}
	return out, rs.Err()
}
