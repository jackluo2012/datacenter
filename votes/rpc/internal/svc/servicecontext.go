package svc

import (
	"datacenter/votes/model"
	"datacenter/votes/rpc/internal/config"

	"github.com/tal-tech/go-zero/core/stores/redis"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	c                     config.Config
	AppEnrollModel        model.AppEnrollModel
	AppVotesActivityModel model.AppVotesActivityModel
	AppVotesModel         model.AppVotesModel
	RedisConn             *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	aem := model.NewAppEnrollModel(conn, c.CacheRedis)
	avm := model.NewAppVotesModel(conn)
	avam := model.NewAppVotesActivityModel(conn, c.CacheRedis)
	rconn := redis.NewRedis(c.CacheRedis[0].Host, c.CacheRedis[0].Type, c.CacheRedis[0].Pass)
	return &ServiceContext{
		c:                     c,
		AppEnrollModel:        aem,
		AppVotesModel:         avm,
		AppVotesActivityModel: avam,
		RedisConn:             rconn,
	}
}
