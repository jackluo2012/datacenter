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

type DeployTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 部署到调度器
func NewDeployTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeployTaskLogic {
	return &DeployTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 部署到调度器: 网关请求逐字段映射为 rpc 请求, 转发后把部署结果转为网关类型返回
func (l *DeployTaskLogic) DeployTask(req *types.SyncTaskIdReq) (resp *types.DeployResp, err error) {
	reply, err := l.svcCtx.JobRpc.DeployToScheduler(l.ctx, &jobClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.DeployResp{
		Ok:             reply.GetOk(),
		DsWorkflowCode: reply.GetDsWorkflowCode(),
		Msg:            reply.GetMsg(),
	}, nil
}
