package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteElementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteElementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteElementLogic {
	return &DeleteElementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteElementLogic) DeleteElement(in *standard.IdReq) (*standard.IdReq, error) {
	// todo: add your logic here and delete this line

	return &standard.IdReq{}, nil
}
