package svc

import (
	"datacenter/user/api/internal/config"
	"datacenter/user/api/internal/middleware"
	"datacenter/user/model"

	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"github.com/tal-tech/go-zero/rest"
)

type ServiceContext struct {
	Config    config.Config
	UserModel *model.BaseMemberModel
	UserCheck rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	um := model.NewBaseMemberModel(conn, c.CacheRedis)

	return &ServiceContext{
		Config:    c,
		UserModel: um,
		UserCheck: middleware.NewUserCheckMiddleware().Handle,
	}
}
