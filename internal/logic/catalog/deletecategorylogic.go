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

type DeleteCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除分类（有子分类或挂有资产时拒绝）
func NewDeleteCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCategoryLogic {
	return &DeleteCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteCategory 转发 catalog rpc 删除分类（软删；引用校验在 rpc 侧）
func (l *DeleteCategoryLogic) DeleteCategory(req *types.CatalogIdReq) error {
	// rpc DeleteCategory 返回 IdReq，网关无返回体，仅透传错误
	_, err := l.svcCtx.CatalogRpc.DeleteCategory(l.ctx, &catalogClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
