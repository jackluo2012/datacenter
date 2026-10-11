package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSyncTaskV2Logic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSyncTaskV2Logic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSyncTaskV2Logic {
	return &CreateSyncTaskV2Logic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建同步任务（内置引擎：源 SQL→目标表 分批同步）
func (l *CreateSyncTaskV2Logic) CreateSyncTaskV2(in *job.SaveSyncTaskReq) (*job.TaskReply, error) {
	// todo: add your logic here and delete this line

	return &job.TaskReply{}, nil
}
