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

type TriggerTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 手动触发任务
func NewTriggerTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TriggerTaskLogic {
	return &TriggerTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 手动触发任务: 网关请求逐字段映射为 rpc 请求, 转发后把触发结果转为网关类型返回
func (l *TriggerTaskLogic) TriggerTask(req *types.SyncTaskIdReq) (resp *types.TriggerResp, err error) {
	reply, err := l.svcCtx.JobRpc.TriggerTask(l.ctx, &jobClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.TriggerResp{
		Ok:                  reply.GetOk(),
		DsProcessInstanceId: reply.GetDsProcessInstanceId(),
		Msg:                 reply.GetMsg(),
	}, nil
}
