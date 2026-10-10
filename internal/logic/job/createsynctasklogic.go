// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package job

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/job/rpc/jobClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSyncTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建同步任务
func NewCreateSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSyncTaskLogic {
	return &CreateSyncTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 创建同步任务: 同步模式默认 offline, 网关请求逐字段映射为 rpc 请求, 转发后把 SyncTask 转为网关类型返回
func (l *CreateSyncTaskLogic) CreateSyncTask(req *types.CreateSyncTaskReq) (resp *types.SyncTask, err error) {
	mode := req.Mode
	if mode == "" {
		mode = "offline"
	}
	reply, err := l.svcCtx.JobRpc.CreateSyncTask(l.ctx, &jobClient.CreateSyncTaskReq{
		Name:        req.Name,
		SourceDsId:  req.SourceDsId,
		TargetDsId:  req.TargetDsId,
		Mode:        mode,
		MappingJson: req.MappingJson,
		Cron:        req.Cron,
		Beid:        req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return syncTaskToTypes(reply.Task), nil
}
