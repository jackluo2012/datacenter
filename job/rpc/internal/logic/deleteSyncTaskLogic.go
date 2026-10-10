package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSyncTaskLogic {
	return &DeleteSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除任务（已部署的先提示需先下线：非草稿状态一律拒绝，避免调度器侧残留工作流）
func (l *DeleteSyncTaskLogic) DeleteSyncTask(in *job.IdReq) (*job.IdReq, error) {
	task, err := loadTask(l.ctx, l.svcCtx.SyncTaskModel, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	// 已部署/运行中/失败的任务在调度器侧仍存在工作流，直接拒绝并提示先下线
	if ok, _ := CanDelete(task.Status); !ok {
		return nil, ErrTaskDeployed
	}
	if err := l.svcCtx.SyncTaskModel.Delete(l.ctx, task.Id, in.Beid); err != nil {
		return nil, err
	}
	return &job.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
