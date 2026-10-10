package svc

import (
	"datacenter/metadata/model"
	"datacenter/metadata/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	Conn          sqlx.SqlConn // 主库连接（跨表事务编排用）
	EntityModel   model.EntityModel
	RelationModel model.RelationModel
	VersionModel  model.VersionModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:        c,
		Conn:          conn,
		EntityModel:   model.NewEntityModel(conn),
		RelationModel: model.NewRelationModel(conn),
		VersionModel:  model.NewVersionModel(conn),
	}
}
