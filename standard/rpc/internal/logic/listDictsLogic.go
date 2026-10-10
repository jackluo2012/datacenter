package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDictsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDictsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDictsLogic {
	return &ListDictsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListDictsLogic) ListDicts(in *standard.DictListReq) (*standard.DictListReply, error) {
	// todo: add your logic here and delete this line

	return &standard.DictListReply{}, nil
}
