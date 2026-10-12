package logic

import (
	"context"

	"datacenter/catalog/model"
	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MoveAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMoveAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MoveAssetLogic {
	return &MoveAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 移动资产（切换/移出分类）：资产存在性校验 → 目标分类存在性校验 → 回写 category_id
func (l *MoveAssetLogic) MoveAsset(in *catalog.MoveAssetReq) (*catalog.AssetReply, error) {
	// 参数校验
	if in.Id <= 0 || in.Beid <= 0 {
		return nil, ErrInvalidParam
	}
	// 资产存在性校验（带平台隔离）
	asset, err := l.svcCtx.AssetModel.FindOne(in.Id, in.Beid)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}
	// 目标分类必校验存在；category_id=0 表示移出分类，落到未分类
	if in.CategoryId > 0 {
		if _, err := l.svcCtx.CategoryModel.FindOne(in.CategoryId, in.Beid); err != nil {
			if err == model.ErrNotFound {
				return nil, ErrCategoryNotFound
			}
			return nil, err
		}
	}
	asset.CategoryId = in.CategoryId
	if err := l.svcCtx.AssetModel.Update(*asset); err != nil {
		l.Errorf("移动资产失败, id: %d, category_id: %d, beid: %d, err: %v", in.Id, in.CategoryId, in.Beid, err)
		return nil, err
	}
	return &catalog.AssetReply{Asset: assetToPb(asset)}, nil
}
