package logic

import (
	"context"

	"datacenter/catalog/model"
	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAssetLogic {
	return &DeleteAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除资产：存在性校验后物理删除；
// 历史申请单（dc_asset_apply）不级联清理，保留审批痕迹供追溯
func (l *DeleteAssetLogic) DeleteAsset(in *catalog.IdReq) (*catalog.IdReq, error) {
	// 参数校验
	if in.Id <= 0 || in.Beid <= 0 {
		return nil, ErrInvalidParam
	}
	// 存在性校验（带平台隔离），避免误删
	if _, err := l.svcCtx.AssetModel.FindOne(in.Id, in.Beid); err != nil {
		if err == model.ErrNotFound {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}
	if err := l.svcCtx.AssetModel.Delete(in.Id, in.Beid); err != nil {
		l.Errorf("删除资产失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, err
	}
	return &catalog.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
