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

// 任务列表（beid 多租户必带 + 分页）
func (l *ListArchiveTasksLogic) ListArchiveTasks(in *ops.TaskListReq) (*ops.TaskListReply, error) {
	list, total, err := l.svcCtx.ArchiveTaskModel.FindList(l.ctx, in.Beid, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	tasks := make([]*ops.ArchiveTask, 0, len(list))
	for _, t := range list {
		tasks = append(tasks, taskToPb(t))
	}
	return &ops.TaskListReply{List: tasks, Total: total}, nil
}
