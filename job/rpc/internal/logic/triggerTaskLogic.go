package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type TriggerTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTriggerTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TriggerTaskLogic {
	return &TriggerTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 手动触发一次
func (l *TriggerTaskLogic) TriggerTask(in *job.IdReq) (*job.TriggerReply, error) {
	// todo: add your logic here and delete this line

	return &job.TriggerReply{}, nil
}
