package logic

import (
	"context"

	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListArchiveTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListArchiveTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListArchiveTasksLogic {
	return &ListArchiveTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 任务列表
func (l *ListArchiveTasksLogic) ListArchiveTasks(in *ops.TaskListReq) (*ops.TaskListReply, error) {
	// todo: add your logic here and delete this line

	return &ops.TaskListReply{}, nil
}
