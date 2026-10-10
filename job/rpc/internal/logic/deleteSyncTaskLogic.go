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

// 删除任务（已部署的先提示需先下线）
func (l *DeleteSyncTaskLogic) DeleteSyncTask(in *job.IdReq) (*job.IdReq, error) {
	// todo: add your logic here and delete this line

	return &job.IdReq{}, nil
}
