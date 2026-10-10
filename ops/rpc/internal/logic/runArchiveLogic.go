package logic

import (
	"context"

	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunArchiveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunArchiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunArchiveLogic {
	return &RunArchiveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行归档（真实跑库：分批 insert-select + delete 事务）
func (l *RunArchiveLogic) RunArchive(in *ops.IdReq) (*ops.RunArchiveReply, error) {
	// todo: add your logic here and delete this line

	return &ops.RunArchiveReply{}, nil
}
