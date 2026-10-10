package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSchedulerInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSchedulerInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSchedulerInfoLogic {
	return &GetSchedulerInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 调度器连接信息（运维视图）
func (l *GetSchedulerInfoLogic) GetSchedulerInfo(in *job.IdReq) (*job.SchedulerReply, error) {
	// todo: add your logic here and delete this line

	return &job.SchedulerReply{}, nil
}
