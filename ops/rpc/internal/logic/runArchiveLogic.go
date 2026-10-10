package logic

import (
	"context"
	"fmt"
	"time"

	"datacenter/datasource/rpc/datasource"
	"datacenter/ops/model"
	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RunArchiveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunArchiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunArchiveLogic {
	return &RunArchiveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行归档（真实跑库）：
// 查任务 → 运行前复校表名/条件 → 数据源服务取 DSN 建目标库连接 →
// 分批 insert-select(+同事务 delete) 循环 → 落 dc_ops_archive_log + 回填任务最近状态。
// 单批失败即停：已提交批保留，失败位置（批号）写入日志 msg，接口仍返回 RunArchiveReply（status=failed）
func (l *RunArchiveLogic) RunArchive(in *ops.IdReq) (*ops.RunArchiveReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, model.ErrTaskIdRequired
	}
	task, err := l.svcCtx.ArchiveTaskModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	// 运行前复校表名/条件（防御历史脏数据直接拼接进 SQL）
	sourceTable, err := QualifyTable(task.DbName, task.TableName)
	if err != nil {
		return nil, err
	}
	archiveTable, err := QualifyTable(task.DbName, task.ArchiveTable)
	if err != nil {
		return nil, err
	}
	if err := ValidateConditionSQL(task.ConditionSql); err != nil {
		return nil, err
	}
	// 目标库连接：task.datasource_id → 数据源服务取 DSN
	conn, err := l.targetConn(task.DatasourceId, in.Beid)
	if err != nil {
		return nil, err
	}
	// 归档表不存在时按源表结构自动创建（like 源表含索引），幂等
	if _, err := conn.ExecCtx(l.ctx,
		"create table if not exists "+archiveTable+" like "+sourceTable); err != nil {
		return nil, fmt.Errorf("创建归档表失败: %w", err)
	}
	batchSize := task.BatchSize
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	// 分批执行：任一批失败即停，失败前已提交批保留
	stats, loopErr := RunArchiveLoop(l.ctx, NewMysqlArchiveDB(conn), ArchiveOptions{
		SourceTable:  sourceTable,
		ArchiveTable: archiveTable,
		Condition:    task.ConditionSql,
		BatchSize:    batchSize,
		DeleteSource: task.DeleteSource == 1,
	})
	status := model.ArchiveStatusSuccess
	msg := ""
	if loopErr != nil {
		status = model.ArchiveStatusFailed
		msg = truncateRunes(fmt.Sprintf("第 %d 批失败: %v", stats.Batches+1, loopErr), archiveMsgMaxLen)
		l.Errorf("归档任务 %d 执行失败: %s", task.Id, msg)
	}
	runAt := time.Now()
	// 执行结果落归档日志（成功失败都落，失败带失败位置）
	if err := l.svcCtx.ArchiveLogModel.Insert(l.ctx, &model.ArchiveLog{
		TaskId:       task.Id,
		ArchivedRows: stats.Archived,
		DeletedRows:  stats.Deleted,
		Status:       status,
		Msg:          msg,
		RunAt:        runAt.UnixMilli(),
	}); err != nil {
		return nil, err
	}
	// 回填任务最近状态
	if err := l.svcCtx.ArchiveTaskModel.UpdateRunStatus(l.ctx, task.Id, in.Beid, status, runAt); err != nil {
		return nil, err
	}
	return &ops.RunArchiveReply{
		ArchivedRows: stats.Archived,
		DeletedRows:  stats.Deleted,
		Status:       status,
	}, nil
}

// targetConn 从数据源服务获取 DSN 并建立目标库连接（with_secret=true 才返回连接串）
func (l *RunArchiveLogic) targetConn(datasourceId, beid int64) (sqlx.SqlConn, error) {
	reply, err := l.svcCtx.DataSourceCli.GetDataSource(l.ctx, &datasource.GetDataSourceReq{
		Id:         datasourceId,
		Beid:       beid,
		WithSecret: true,
	})
	if err != nil {
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}
	if reply == nil || reply.Datasource == nil || reply.Dsn == "" {
		return nil, ErrDataSourceDsn
	}
	if reply.Datasource.Type != "mysql" {
		return nil, fmt.Errorf("%w: 归档仅支持 mysql 数据源，当前 %q", ErrDataSourceDsn, reply.Datasource.Type)
	}
	// DSN 已含默认库；task.db_name 非空时由 QualifyTable 在表名前加 `<db>.` 前缀
	return sqlx.NewMysql(reply.Dsn), nil
}
