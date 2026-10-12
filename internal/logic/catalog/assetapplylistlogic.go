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

type AssetApplyListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 资产申请列表
func NewAssetApplyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssetApplyListLogic {
	return &AssetApplyListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AssetApplyList 转发 catalog rpc 查询申请分页列表（status 空查全部）
func (l *AssetApplyListLogic) AssetApplyList(req *types.AssetApplyListReq) (resp *types.AssetApplyListResp, err error) {
	// 分页默认值：page<=0 取 1，size<=0 取 10
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 10
	}
	rpcResp, err := l.svcCtx.CatalogRpc.ListApplies(l.ctx, &catalogClient.ApplyListReq{
		Beid:   req.Beid,
		Status: req.Status,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}
	// 逐条映射申请列表（make 保证空列表序列化为 [] 而非 null）
	list := make([]types.AssetApplyInfo, 0, len(rpcResp.GetList()))
	for _, a := range rpcResp.GetList() {
		list = append(list, toAssetApply(a))
	}
	return &types.AssetApplyListResp{List: list, Total: rpcResp.GetTotal()}, nil
}
