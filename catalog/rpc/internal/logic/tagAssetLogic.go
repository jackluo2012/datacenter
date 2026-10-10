package logic

import (
	"context"

	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TagAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTagAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagAssetLogic {
	return &TagAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登记资产（元数据实体入目录）
func (l *TagAssetLogic) TagAsset(in *catalog.TagAssetReq) (*catalog.AssetReply, error) {
	// todo: add your logic here and delete this line

	return &catalog.AssetReply{}, nil
}
