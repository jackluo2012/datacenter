package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/job/model"
	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RunSyncLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunSyncLogic {
	return &RunSyncLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行同步（内置引擎真实跑库：分批读源写目标，每批一个事务）：
// 查任务 → 运行前复校（draft/deployed 等任意状态均可跑）→ 双 DSN 装配
// （source_ds_id + target_ds_id 走数据源服务 with_secret；target_ds_id=0 目标为平台自有库）→
// 引擎分批同步 → 落 dc_sync_task_log + 回填任务状态 → RunSyncReply。
// 任务加载后的一切执行失败（校验/连接/映射/同步）不返回 error：status=failed + msg
// （含失败批号），已提交批保留，与 ops RunArchive 同语义
func (l *RunSyncLogic) RunSync(in *job.RunSyncReq) (*job.RunSyncReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	// 任务不存在/跨租户不可见：无 task_id 可落日志，返回 error
	task, err := loadTask(l.ctx, l.svcCtx.SyncTaskModel, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	// 准备阶段（静态复校 + 目标表全限定名 + 列映射 + 双库连接），任一失败按 failed 收口
	qualifiedTable, mapping, srcDB, tgtDB, prepErr := l.prepare(task, in.Beid)
	var stats RunStats
	var loopErr error
	if prepErr == nil {
		// 引擎分批主循环（mode=cdc 本期按 offline 语义执行：全量分批快照，无位点管理）
		stats, loopErr = RunSyncLoop(l.ctx, srcDB, tgtDB, SyncOptions{
			SourceSql:      task.SourceSql,
			QualifiedTable: qualifiedTable,
			Mapping:        mapping,
			BatchSize:      task.BatchSize,
		})
	}
	// 运行终态：失败 msg 统一带失败批号（失败批号 = 已提交批数 + 1）
	status := model.SyncLogStatusSuccess
	msg := ""
	switch {
	case prepErr != nil:
		status = model.SyncLogStatusFailed
		msg = truncateRunes(prepErr.Error(), syncMsgMaxLen)
	case loopErr != nil:
		status = model.SyncLogStatusFailed
		msg = truncateRunes(fmt.Sprintf("第 %d 批失败: %v", stats.Batches+1, loopErr), syncMsgMaxLen)
		l.Errorf("内置同步任务 %d 执行失败: %s", task.Id, msg)
	}
	costMs := time.Since(start).Milliseconds()
	// 执行结果落 dc_sync_task_log（成功失败都落；观测面失败仅记日志，不吞执行结果）
	if err := l.svcCtx.SyncLogModel.Insert(l.ctx, &model.SyncLog{
		TaskId:    task.Id,
		ReadRows:  stats.ReadRows,
		WriteRows: stats.WriteRows,
		Status:    status,
		Msg:       msg,
		RunAt:     start.UnixMilli(),
	}); err != nil {
		l.Errorf("同步日志落库失败: task=%d err=%v", task.Id, err)
	}
	// 任务状态回填：失败→failed；成功→草稿保持草稿、其余回落 deployed
	if err := l.svcCtx.SyncTaskModel.UpdateRunStatus(l.ctx, task.Id, in.Beid,
		NextTaskStatus(task.Status, status), time.Now()); err != nil {
		l.Errorf("任务状态回填失败: task=%d err=%v", task.Id, err)
	}
	if status == model.SyncLogStatusSuccess {
		l.Infof("内置同步任务完成: task=%d read=%d write=%d batches=%d cost=%dms",
			task.Id, stats.ReadRows, stats.WriteRows, stats.Batches, costMs)
	}
	return &job.RunSyncReply{
		ReadRows:  stats.ReadRows,
		WriteRows: stats.WriteRows,
		Status:    status,
		Msg:       msg,
		CostMs:    costMs,
	}, nil
}

// prepare 运行前准备：静态复校 + 目标表全限定名 + 列映射解析 + 双库连接装配。
// 返回 error 时由调用方按执行失败收口（落 failed 日志），不返回 error 给网关
func (l *RunSyncLogic) prepare(task *model.SyncTask, beid int64) (string, []ColumnMapping, SyncSourceDB, SyncTargetDB, error) {
	// 运行前复校（与 CreateSyncTaskV2 同一闸门，防御历史脏数据直入引擎）
	if err := ValidateSyncTaskV2(task.Name, task.Mode, task.SourceDsId,
		task.SourceSql, task.TargetTable, task.MappingJson); err != nil {
		return "", nil, nil, nil, err
	}
	// 目标表全限定名（target_db 空 = 目标 DSN 默认库）
	qualifiedTable, err := QualifyTable(strings.TrimSpace(task.TargetDb), strings.TrimSpace(task.TargetTable))
	if err != nil {
		return "", nil, nil, nil, err
	}
	// 列映射解析（空 = 按源结果集列名直通）
	mapping, err := ParseMapping(task.MappingJson)
	if err != nil {
		return "", nil, nil, nil, err
	}
	// 双 DSN：源数据源必配；target_ds_id=0 时目标为平台自有库（本服务主库）
	sourceConn, err := l.dsConn(task.SourceDsId, beid, "源")
	if err != nil {
		return "", nil, nil, nil, err
	}
	var targetConn sqlx.SqlConn
	if task.TargetDsId > 0 {
		if targetConn, err = l.dsConn(task.TargetDsId, beid, "目标"); err != nil {
			return "", nil, nil, nil, err
		}
	} else {
		targetConn = l.svcCtx.Conn
	}
	srcDB, err := NewMysqlSourceDB(sourceConn)
	if err != nil {
		return "", nil, nil, nil, err
	}
	return qualifiedTable, mapping, srcDB, NewMysqlTargetDB(targetConn), nil
}

// dsConn 从数据源服务取 DSN 建连接（with_secret=true 才返回连接串，
// 参考 quality RunJob 的数据源装配模式；role 用于错误信息区分源/目标）
func (l *RunSyncLogic) dsConn(datasourceId, beid int64, role string) (sqlx.SqlConn, error) {
	reply, err := l.svcCtx.DataSourceCli.GetDataSource(l.ctx, &dataSourceClient.GetDataSourceReq{
		Id:         datasourceId,
		Beid:       beid,
		WithSecret: true,
	})
	if err != nil {
		return nil, fmt.Errorf("获取%s数据源失败: %w", role, err)
	}
	if reply == nil || reply.Datasource == nil || reply.Dsn == "" {
		return nil, fmt.Errorf("%w: %s数据源 %d", ErrDataSourceDsn, role, datasourceId)
	}
	if reply.Datasource.Type != "mysql" {
		return nil, fmt.Errorf("%w: 内置引擎仅支持 mysql 数据源，%s数据源 %d 为 %q",
			ErrDataSourceDsn, role, datasourceId, reply.Datasource.Type)
	}
	// DSN 已含默认库；task.target_db 非空时由 QualifyTable 在表名前加 `<db>.` 前缀
	return sqlx.NewMysql(reply.Dsn), nil
}
