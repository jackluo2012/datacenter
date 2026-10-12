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

type ReviewAssetApplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 审批资产申请
func NewReviewAssetApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewAssetApplyLogic {
	return &ReviewAssetApplyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ReviewAssetApply 转发 catalog rpc 审批申请：仅 pending 可审，approve 决定通过/驳回
// （状态机与防并发重复审批在 rpc 侧）
func (l *ReviewAssetApplyLogic) ReviewAssetApply(req *types.ReviewAssetApplyReq) (resp *types.AssetApplyInfo, err error) {
	rpcResp, err := l.svcCtx.CatalogRpc.ReviewApply(l.ctx, &catalogClient.ReviewApplyReq{
		Id:       req.Id,
		Approve:  req.Approve,
		Reviewer: req.Reviewer,
		Beid:     req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}
	apply := toAssetApply(rpcResp.GetApply())
	return &apply, nil
}
