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

type CreateCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建资产分类
func NewCreateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCategoryLogic {
	return &CreateCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateCategory 转发 catalog rpc 创建分类，返回含 id 的 AssetCategory
func (l *CreateCategoryLogic) CreateCategory(req *types.CreateCategoryReq) (resp *types.AssetCategory, err error) {
	// rpc 入参即 AssetCategory：逐字段映射（网关契约字段是它的子集）
	rpcResp, err := l.svcCtx.CatalogRpc.CreateCategory(l.ctx, &catalogClient.AssetCategory{
		ParentId: req.ParentId,
		Name:     req.Name,
		Beid:     req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	// 逐字段映射 rpc 返回的分类（含生成的 id）
	category := toCategory(rpcResp.GetCategory())
	return &category, nil
}
