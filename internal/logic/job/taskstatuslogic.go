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

type TaskStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 任务状态
func NewTaskStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskStatusLogic {
	return &TaskStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 任务状态: 网关请求逐字段映射为 rpc 请求, 转发后把状态聚合结果转为网关类型返回
func (l *TaskStatusLogic) TaskStatus(req *types.SyncTaskIdReq) (resp *types.TaskStatusResp, err error) {
	reply, err := l.svcCtx.JobRpc.GetTaskStatus(l.ctx, &jobClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.TaskStatusResp{
		Status:       reply.GetStatus(),
		DsInstanceId: reply.GetDsInstanceId(),
		Msg:          reply.GetMsg(),
	}, nil
}
