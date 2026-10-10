// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package catalog

import (
	"context"

	"datacenter/catalog/rpc/catalogClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TagAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 登记资产
func NewTagAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagAssetLogic {
	return &TagAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// TagAsset 转发 catalog rpc 登记资产（元数据实体入目录），返回登记后的 Asset
func (l *TagAssetLogic) TagAsset(req *types.TagAssetReq) (resp *types.Asset, err error) {
	// classification 未传时默认 L2（分级：L1/L2/L3）
	classification := req.Classification
	if classification == "" {
		classification = "L2"
	}
	// visibility 未传时默认 private（public/private）
	visibility := req.Visibility
	if visibility == "" {
		visibility = "private"
	}

	// 逐字段映射为 rpc 请求
	rpcResp, err := l.svcCtx.CatalogRpc.TagAsset(l.ctx, &catalogClient.TagAssetReq{
		EntityId:       req.EntityId,
		CategoryId:     req.CategoryId,
		Name:           req.Name,
		EntityType:     req.EntityType,
		Comment:        req.Comment,
		Classification: classification,
		AssetTagsJson:  req.AssetTagsJson,
		Visibility:     visibility,
		Beid:           req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	// 逐字段映射 rpc 返回的资产（网关 Asset 契约无 beid 字段，不透出）
	a := rpcResp.GetAsset()
	return &types.Asset{
		Id:             a.GetId(),
		EntityId:       a.GetEntityId(),
		CategoryId:     a.GetCategoryId(),
		Name:           a.GetName(),
		EntityType:     a.GetEntityType(),
		Comment:        a.GetComment(),
		Classification: a.GetClassification(),
		AssetTagsJson:  a.GetAssetTagsJson(),
		Visibility:     a.GetVisibility(),
	}, nil
}
