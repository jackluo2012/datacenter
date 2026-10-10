package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTaskStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskStatusLogic {
	return &GetTaskStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 任务状态（本地 + 调度器实例状态）
func (l *GetTaskStatusLogic) GetTaskStatus(in *job.IdReq) (*job.StatusReply, error) {
	// todo: add your logic here and delete this line

	return &job.StatusReply{}, nil
}
