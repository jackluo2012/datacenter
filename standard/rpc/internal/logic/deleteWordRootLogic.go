package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteWordRootLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteWordRootLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteWordRootLogic {
	return &DeleteWordRootLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteWordRootLogic) DeleteWordRoot(in *standard.IdReq) (*standard.IdReq, error) {
	// todo: add your logic here and delete this line

	return &standard.IdReq{}, nil
}
