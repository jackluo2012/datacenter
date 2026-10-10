package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteModelLogic {
	return &DeleteModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteModelLogic) DeleteModel(in *standard.IdReq) (*standard.IdReq, error) {
	// todo: add your logic here and delete this line

	return &standard.IdReq{}, nil
}
