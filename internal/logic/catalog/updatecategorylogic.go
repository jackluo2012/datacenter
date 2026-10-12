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

type UpdateCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新分类
func NewUpdateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCategoryLogic {
	return &UpdateCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateCategory 转发 catalog rpc 更新分类（改名/挂父，rpc 侧做成环与存在性校验）
func (l *UpdateCategoryLogic) UpdateCategory(req *types.UpdateCategoryReq) (resp *types.AssetCategory, err error) {
	rpcResp, err := l.svcCtx.CatalogRpc.UpdateCategory(l.ctx, &catalogClient.UpdateCategoryReq{
		Id:       req.Id,
		ParentId: req.ParentId,
		Name:     req.Name,
		Beid:     req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}
	category := toCategory(rpcResp.GetCategory())
	return &category, nil
}
