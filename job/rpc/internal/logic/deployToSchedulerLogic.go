package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeployToSchedulerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeployToSchedulerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeployToSchedulerLogic {
	return &DeployToSchedulerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 部署到调度器（DolphinScheduler；未配置/不可达时 ok=false 带原因）
func (l *DeployToSchedulerLogic) DeployToScheduler(in *job.IdReq) (*job.DeployReply, error) {
	// todo: add your logic here and delete this line

	return &job.DeployReply{}, nil
}
