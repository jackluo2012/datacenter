package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveWordRootLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveWordRootLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveWordRootLogic {
	return &SaveWordRootLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 词根
func (l *SaveWordRootLogic) SaveWordRoot(in *standard.SaveWordRootReq) (*standard.WordRootReply, error) {
	// todo: add your logic here and delete this line

	return &standard.WordRootReply{}, nil
}
