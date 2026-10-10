package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListElementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListElementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListElementsLogic {
	return &ListElementsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListElementsLogic) ListElements(in *standard.ElementListReq) (*standard.ElementListReply, error) {
	// todo: add your logic here and delete this line

	return &standard.ElementListReply{}, nil
}
