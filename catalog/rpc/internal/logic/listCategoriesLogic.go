package logic

import (
	"context"

	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCategoriesLogic {
	return &ListCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分类树（按 parent_id 升序平铺返回，前端按 parent_id 组树）
func (l *ListCategoriesLogic) ListCategories(in *catalog.IdReq) (*catalog.CategoryListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrInvalidParam
	}
	list, err := l.svcCtx.CategoryModel.ListByBeid(in.Beid)
	if err != nil {
		return nil, err
	}
	reply := make([]*catalog.AssetCategory, 0, len(list))
	for _, c := range list {
		reply = append(reply, &catalog.AssetCategory{
			Id:       c.Id,
			ParentId: c.ParentId,
			Name:     c.Name,
			Beid:     c.Beid,
		})
	}
	return &catalog.CategoryListReply{List: reply}, nil
}
