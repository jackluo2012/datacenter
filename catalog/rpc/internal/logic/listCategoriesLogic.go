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

// 分类树
func (l *ListCategoriesLogic) ListCategories(in *catalog.IdReq) (*catalog.CategoryListReply, error) {
	// todo: add your logic here and delete this line

	return &catalog.CategoryListReply{}, nil
}
