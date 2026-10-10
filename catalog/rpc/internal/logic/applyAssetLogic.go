package logic

import (
	"context"

	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApplyAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyAssetLogic {
	return &ApplyAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 申请使用
func (l *ApplyAssetLogic) ApplyAsset(in *catalog.ApplyAssetReq) (*catalog.ApplyReply, error) {
	// todo: add your logic here and delete this line

	return &catalog.ApplyReply{}, nil
}
