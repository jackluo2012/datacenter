package logic

import (
	"context"

	"datacenter/catalog/model"
	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCategoryLogic {
	return &CreateCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建目录分类
func (l *CreateCategoryLogic) CreateCategory(in *catalog.AssetCategory) (*catalog.CategoryReply, error) {
	// 参数校验：名称必填
	if in.Name == "" || in.Beid <= 0 {
		return nil, ErrInvalidParam
	}
	// 指定父分类时校验其存在（带平台隔离）
	if in.ParentId > 0 {
		if _, err := l.svcCtx.CategoryModel.FindOne(in.ParentId, in.Beid); err != nil {
			if err == model.ErrNotFound {
				return nil, ErrCategoryNotFound
			}
			return nil, err
		}
	}
	ret, err := l.svcCtx.CategoryModel.Insert(model.AssetCategory{
		ParentId: in.ParentId,
		Name:     in.Name,
		Beid:     in.Beid,
	})
	if err != nil {
		return nil, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &catalog.CategoryReply{
		Category: &catalog.AssetCategory{
			Id:       id,
			ParentId: in.ParentId,
			Name:     in.Name,
			Beid:     in.Beid,
		},
	}, nil
}
