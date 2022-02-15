package svc

import (
	"datacenter/questions/model"
	"datacenter/questions/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                   config.Config
	RedisConn                *redis.Redis
	QuestionsActivitiesModel model.AppQuestionsActivitiesModel
	QuestionsAwardsModel     model.AppQuestionsAwardsModel
	QuestionsModel           model.AppQuestionsModel
	QuestionsAnswersModel    model.AppQuestionsAnswersModel
	QuestionsLotteriesModel  model.AppQuestionsLotteriesModel
	QuestionsConvertsModel   model.AppQuestionsConvertsModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	rconn := redis.NewRedis(c.CacheRedis[0].Host, c.CacheRedis[0].Type, c.CacheRedis[0].Pass)
	mconn := sqlx.NewMysql(c.Mysql.DataSource)
	qam := model.NewAppQuestionsActivitiesModel(mconn, c.CacheRedis)
	qawardsm := model.NewAppQuestionsAwardsModel(mconn, c.CacheRedis)
	qm := model.NewAppQuestionsModel(mconn, c.CacheRedis)
	qanswersm := model.NewAppQuestionsAnswersModel(mconn, c.CacheRedis)
	qlm := model.NewAppQuestionsLotteriesModel(mconn, c.CacheRedis)
	qcm := model.NewAppQuestionsConvertsModel(mconn, c.CacheRedis)
	return &ServiceContext{
		Config:                   c,
		RedisConn:                rconn,
		QuestionsActivitiesModel: qam,
		QuestionsAwardsModel:     qawardsm,
		QuestionsModel:           qm,
		QuestionsAnswersModel:    qanswersm,
		QuestionsLotteriesModel:  qlm,
		QuestionsConvertsModel:   qcm,
	}
}
