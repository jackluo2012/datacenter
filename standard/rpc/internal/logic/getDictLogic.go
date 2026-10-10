package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictLogic {
	return &GetDictLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDictLogic) GetDict(in *standard.IdReq) (*standard.DictReply, error) {
	// todo: add your logic here and delete this line

	return &standard.DictReply{}, nil
}
