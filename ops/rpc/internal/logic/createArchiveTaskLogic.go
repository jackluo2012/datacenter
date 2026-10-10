package logic

import (
	"context"

	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateArchiveTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateArchiveTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateArchiveTaskLogic {
	return &CreateArchiveTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建归档任务
func (l *CreateArchiveTaskLogic) CreateArchiveTask(in *ops.CreateArchiveTaskReq) (*ops.TaskReply, error) {
	// todo: add your logic here and delete this line

	return &ops.TaskReply{}, nil
}
