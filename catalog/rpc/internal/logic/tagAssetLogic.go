package logic

import (
	"context"

	"datacenter/catalog/model"
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

// 登记资产（元数据实体入目录，uk_beid_entity 冲突则刷新快照）
func (l *TagAssetLogic) TagAsset(in *catalog.TagAssetReq) (*catalog.AssetReply, error) {
	// 参数校验
	if in.EntityId <= 0 || in.Name == "" || in.Beid <= 0 {
		return nil, ErrInvalidParam
	}
	// 指定分类时校验其存在（带平台隔离）
	if in.CategoryId > 0 {
		if _, err := l.svcCtx.CategoryModel.FindOne(in.CategoryId, in.Beid); err != nil {
			if err == model.ErrNotFound {
				return nil, ErrCategoryNotFound
			}
			return nil, err
		}
	}
	// 可见性默认 private
	visibility := in.Visibility
	if visibility == "" {
		visibility = VisibilityPrivate
	}
	id, err := l.svcCtx.AssetModel.TagAsset(model.Asset{
		EntityId:       in.EntityId,
		CategoryId:     in.CategoryId,
		Name:           in.Name,
		EntityType:     in.EntityType,
		Comment:        in.Comment,
		Classification: in.Classification,
		AssetTagsJson:  in.AssetTagsJson,
		Visibility:     visibility,
		Beid:           in.Beid,
	})
	if err != nil {
		return nil, err
	}
	// 回查落库后的完整快照（含创建/更新时间语义由冗余快照体现）
	asset, err := l.svcCtx.AssetModel.FindOne(id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &catalog.AssetReply{Asset: assetToPb(asset)}, nil
}

// assetToPb model 转 pb
func assetToPb(a *model.Asset) *catalog.Asset {
	return &catalog.Asset{
		Id:             a.Id,
		EntityId:       a.EntityId,
		CategoryId:     a.CategoryId,
		Name:           a.Name,
		EntityType:     a.EntityType,
		Comment:        a.Comment,
		Classification: a.Classification,
		AssetTagsJson:  a.AssetTagsJson,
		Visibility:     a.Visibility,
		Beid:           a.Beid,
	}
}
