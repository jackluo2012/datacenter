package svc

import (
	"datacenter/common/model"
	"datacenter/common/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	c              config.Config
	AppConfigModel model.AppConfigModel
	BaseAppModel   model.BaseAppModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	apm := model.NewAppConfigModel(conn, c.CacheRedis)
	bam := model.NewBaseAppModel(conn, c.CacheRedis)
	return &ServiceContext{
		c:              c,
		AppConfigModel: apm,
		BaseAppModel:   bam,
	}
}
