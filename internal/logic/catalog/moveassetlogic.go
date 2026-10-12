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

type MoveAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 移动资产（切换/移出分类）
func NewMoveAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MoveAssetLogic {
	return &MoveAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MoveAsset 转发 catalog rpc 移动资产：category_id>0 切换分类，=0 移出分类（校验在 rpc 侧）
func (l *MoveAssetLogic) MoveAsset(req *types.MoveAssetReq) (resp *types.Asset, err error) {
	rpcResp, err := l.svcCtx.CatalogRpc.MoveAsset(l.ctx, &catalogClient.MoveAssetReq{
		Id:         req.Id,
		CategoryId: req.CategoryId,
		Beid:       req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}
	asset := toAsset(rpcResp.GetAsset())
	return &asset, nil
}
