package logic

import (
	"context"

	"datacenter/catalog/model"
	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCategoryLogic {
	return &DeleteCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除分类（软删）：存在性校验 → 有子分类拒删 → 分类下挂有资产拒删 → 软删落库
func (l *DeleteCategoryLogic) DeleteCategory(in *catalog.IdReq) (*catalog.IdReq, error) {
	// 参数校验
	if in.Id <= 0 || in.Beid <= 0 {
		return nil, ErrInvalidParam
	}
	// 存在性校验（带平台隔离，软删的查不到即视为不存在）
	if _, err := l.svcCtx.CategoryModel.FindOne(in.Id, in.Beid); err != nil {
		if err == model.ErrNotFound {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	// 引用校验一：存在子分类时拒绝删除，避免子分类变孤儿
	children, err := l.svcCtx.CategoryModel.CountChildren(in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	if children > 0 {
		return nil, ErrCategoryHasChildren
	}
	// 引用校验二：分类下挂有资产时拒绝删除
	assets, err := l.svcCtx.AssetModel.CountByCategory(in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	if assets > 0 {
		return nil, ErrCategoryHasAssets
	}
	if err := l.svcCtx.CategoryModel.Delete(in.Id, in.Beid); err != nil {
		l.Errorf("删除分类失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, err
	}
	return &catalog.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
