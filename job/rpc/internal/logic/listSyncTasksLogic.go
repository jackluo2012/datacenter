package logic

import (
	"context"

	"datacenter/job/model"
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

// 任务列表（分页 + status 可选过滤）
func (l *ListSyncTasksLogic) ListSyncTasks(in *job.TaskListReq) (*job.TaskListReply, error) {
	// 多租户条件必填
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	// status 过滤入参白名单校验（允许空 = 不过滤）
	if in.Status != "" && !model.IsValidTaskStatus(in.Status) {
		return nil, ErrInvalidStatus
	}
	list, total, err := l.svcCtx.SyncTaskModel.FindList(l.ctx, in.Beid, in.Status, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	tasks := make([]*job.SyncTask, 0, len(list))
	for _, t := range list {
		tasks = append(tasks, taskToPb(t))
	}
	return &job.TaskListReply{List: tasks, Total: total}, nil
}
