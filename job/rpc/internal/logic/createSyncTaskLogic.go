package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSyncTaskLogic {
	return &CreateSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建同步任务
func (l *CreateSyncTaskLogic) CreateSyncTask(in *job.CreateSyncTaskReq) (*job.TaskReply, error) {
	// todo: add your logic here and delete this line

	return &job.TaskReply{}, nil
}
