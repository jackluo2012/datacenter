package logic

import (
	"context"

	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAssetsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAssetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAssetsLogic {
	return &ListAssetsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 资产列表
func (l *ListAssetsLogic) ListAssets(in *catalog.AssetListReq) (*catalog.AssetListReply, error) {
	// todo: add your logic here and delete this line

	return &catalog.AssetListReply{}, nil
}
