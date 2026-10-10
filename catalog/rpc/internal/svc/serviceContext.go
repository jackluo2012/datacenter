package svc

import (
	"datacenter/catalog/model"
	"datacenter/catalog/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	CategoryModel model.CategoryModel // 资产目录分类
	AssetModel    model.AssetModel    // 资产
	ApplyModel    model.ApplyModel    // 资产申请使用
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 无缓存需求的表统一用 sqlx 直连
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:        c,
		CategoryModel: model.NewCategoryModel(conn),
		AssetModel:    model.NewAssetModel(conn),
		ApplyModel:    model.NewApplyModel(conn),
	}
}
