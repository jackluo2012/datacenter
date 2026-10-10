package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListWordRootsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWordRootsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWordRootsLogic {
	return &ListWordRootsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListWordRootsLogic) ListWordRoots(in *standard.WordRootListReq) (*standard.WordRootListReply, error) {
	// todo: add your logic here and delete this line

	return &standard.WordRootListReply{}, nil
}
