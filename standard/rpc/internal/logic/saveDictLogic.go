package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveDictLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveDictLogic {
	return &SaveDictLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 字典（含项，整存整取）
func (l *SaveDictLogic) SaveDict(in *standard.SaveDictReq) (*standard.DictReply, error) {
	// todo: add your logic here and delete this line

	return &standard.DictReply{}, nil
}
