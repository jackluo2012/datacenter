// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package ops

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/ops/rpc/opsClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateArchiveTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建归档任务
func NewCreateArchiveTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateArchiveTaskLogic {
	return &CreateArchiveTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 创建归档任务: 批大小缺省 1000(batch_size<=0 取 1000), DeleteSource 布尔直传,
// 网关请求逐字段映射为 rpc 请求, 转发后把任务转为网关类型返回, 业务错误直接透传
func (l *CreateArchiveTaskLogic) CreateArchiveTask(req *types.CreateArchiveTaskReq) (resp *types.OpsArchiveTask, err error) {
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}
	reply, err := l.svcCtx.OpsRpc.CreateArchiveTask(l.ctx, &opsClient.CreateArchiveTaskReq{
		Name:         req.Name,
		DatasourceId: req.DatasourceId,
		DbName:       req.DbName,
		TableName:    req.TableName,
		ConditionSql: req.ConditionSql,
		ArchiveTable: req.ArchiveTable,
		DeleteSource: req.DeleteSource,
		BatchSize:    batchSize,
		Cron:         req.Cron,
		Beid:         req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return archiveTaskToTypes(reply.GetTask()), nil
}
