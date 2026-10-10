package logic

import (
	"context"

	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteArchiveTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteArchiveTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteArchiveTaskLogic {
	return &DeleteArchiveTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除任务
func (l *DeleteArchiveTaskLogic) DeleteArchiveTask(in *ops.IdReq) (*ops.IdReq, error) {
	// todo: add your logic here and delete this line

	return &ops.IdReq{}, nil
}
