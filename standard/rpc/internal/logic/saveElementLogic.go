package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveElementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveElementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveElementLogic {
	return &SaveElementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 数据元
func (l *SaveElementLogic) SaveElement(in *standard.SaveElementReq) (*standard.ElementReply, error) {
	// todo: add your logic here and delete this line

	return &standard.ElementReply{}, nil
}
