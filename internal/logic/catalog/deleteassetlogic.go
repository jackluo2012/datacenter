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

type DeleteAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除资产
func NewDeleteAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAssetLogic {
	return &DeleteAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteAsset 转发 catalog rpc 删除资产（存在性校验在 rpc 侧）
func (l *DeleteAssetLogic) DeleteAsset(req *types.CatalogIdReq) error {
	// rpc DeleteAsset 返回 IdReq，网关无返回体，仅透传错误
	_, err := l.svcCtx.CatalogRpc.DeleteAsset(l.ctx, &catalogClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
