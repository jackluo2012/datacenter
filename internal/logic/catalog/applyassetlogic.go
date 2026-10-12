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

type ApplyAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 申请使用资产
func NewApplyAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyAssetLogic {
	return &ApplyAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ApplyAsset 转发 catalog rpc 提交资产使用申请（落 pending 申请单，资产存在性校验在 rpc 侧）
func (l *ApplyAssetLogic) ApplyAsset(req *types.AssetApplyReq) (resp *types.AssetApplyInfo, err error) {
	rpcResp, err := l.svcCtx.CatalogRpc.ApplyAsset(l.ctx, &catalogClient.ApplyAssetReq{
		AssetId:   req.AssetId,
		Reason:    req.Reason,
		Applicant: req.Applicant,
		Beid:      req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}
	apply := toAssetApply(rpcResp.GetApply())
	return &apply, nil
}
