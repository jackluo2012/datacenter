package svc

import (
	"datacenter/datasource/model"
	"datacenter/datasource/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ServiceContext 服务依赖装配
type ServiceContext struct {
	Config          config.Config
	DataSourceModel model.DataSourceModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	// mysql 连接池, 全服务共用
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:          c,
		DataSourceModel: model.NewDataSourceModel(conn),
	}
}
