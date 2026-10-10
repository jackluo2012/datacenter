package logic

import (
	"context"

	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type OfflineApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOfflineApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OfflineApiLogic {
	return &OfflineApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OfflineApiLogic) OfflineApi(in *datasvc.IdReq) (*datasvc.OkReply, error) {
	// todo: add your logic here and delete this line

	return &datasvc.OkReply{}, nil
}
