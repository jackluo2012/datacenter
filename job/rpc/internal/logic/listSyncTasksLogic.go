package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSyncTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSyncTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSyncTasksLogic {
	return &ListSyncTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 任务列表
func (l *ListSyncTasksLogic) ListSyncTasks(in *job.TaskListReq) (*job.TaskListReply, error) {
	// todo: add your logic here and delete this line

	return &job.TaskListReply{}, nil
}
