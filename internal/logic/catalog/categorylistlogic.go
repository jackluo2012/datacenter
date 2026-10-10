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

type CategoryListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 分类列表
func NewCategoryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CategoryListLogic {
	return &CategoryListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CategoryList 转发 catalog rpc 查询分类列表
func (l *CategoryListLogic) CategoryList(req *types.CategoryListReq) (resp *types.CategoryListResp, err error) {
	// rpc ListCategories 入参为 IdReq：网关只传 beid 做租户过滤，
	// id 留 0 表示取该 beid 下的全部分类（分类树）
	rpcResp, err := l.svcCtx.CatalogRpc.ListCategories(l.ctx, &catalogClient.IdReq{
		Beid: req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	// 逐条映射分类列表（make 保证空列表序列化为 [] 而非 null）
	list := make([]types.AssetCategory, 0, len(rpcResp.GetList()))
	for _, c := range rpcResp.GetList() {
		list = append(list, toCategory(c))
	}

	return &types.CategoryListResp{List: list}, nil
}
