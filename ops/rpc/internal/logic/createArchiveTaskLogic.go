package logic

import (
	"context"
	"fmt"
	"strings"

	"datacenter/ops/model"
	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateArchiveTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateArchiveTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateArchiveTaskLogic {
	return &CreateArchiveTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建归档任务：必填校验 + 库/表名标识符白名单 + condition_sql 黑名单校验 +
// batch_size 缺省兜底，全部通过才落库
func (l *CreateArchiveTaskLogic) CreateArchiveTask(in *ops.CreateArchiveTaskReq) (*ops.TaskReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, ErrTaskNameRequired
	}
	if in.DatasourceId <= 0 {
		return nil, ErrDatasourceRequired
	}
	if strings.TrimSpace(in.TableName) == "" {
		return nil, ErrTableNameRequired
	}
	if strings.TrimSpace(in.ArchiveTable) == "" {
		return nil, ErrArchiveTableRequired
	}
	// 库名/表名要拼接进归档 SQL，先过标识符白名单（防 `t; drop table x` 类注入）
	if in.DbName != "" && !ValidIdentifier(in.DbName) {
		return nil, fmt.Errorf("%w: db_name %q", ErrInvalidIdentifier, in.DbName)
	}
	if !ValidIdentifier(in.TableName) {
		return nil, fmt.Errorf("%w: table_name %q", ErrInvalidIdentifier, in.TableName)
	}
	if !ValidIdentifier(in.ArchiveTable) {
		return nil, fmt.Errorf("%w: archive_table %q", ErrInvalidIdentifier, in.ArchiveTable)
	}
	// 源表与归档表相同会导致 insert select 自循环，直接拒绝
	if in.TableName == in.ArchiveTable {
		return nil, ErrSameTable
	}
	// 归档条件：以 where 开头或为空，且不在黑名单
	if err := ValidateConditionSQL(in.ConditionSql); err != nil {
		return nil, err
	}
	batchSize := int64(in.BatchSize)
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	id, err := l.svcCtx.ArchiveTaskModel.Insert(l.ctx, &model.ArchiveTask{
		Name:         in.Name,
		DatasourceId: in.DatasourceId,
		DbName:       in.DbName,
		TableName:    in.TableName,
		ConditionSql: strings.TrimSpace(in.ConditionSql),
		ArchiveTable: in.ArchiveTable,
		DeleteSource: b2i(in.DeleteSource),
		BatchSize:    batchSize,
		Cron:         in.Cron,
		Beid:         in.Beid,
	})
	if err != nil {
		return nil, err
	}
	task, err := l.svcCtx.ArchiveTaskModel.FindOne(l.ctx, id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &ops.TaskReply{Task: taskToPb(task)}, nil
}

// b2i bool -> tinyint（1/0）
func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
