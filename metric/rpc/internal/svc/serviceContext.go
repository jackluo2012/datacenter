package svc

import (
	"datacenter/metric/model"
	"datacenter/metric/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	Conn           sqlx.SqlConn // 主库连接（跨表事务编排用：级联删 rel/覆盖绑定）
	MetricModel    model.MetricModel
	MetricRelModel model.MetricRelModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:         c,
		Conn:           conn,
		MetricModel:    model.NewMetricModel(conn),
		MetricRelModel: model.NewMetricRelModel(conn),
	}
}
