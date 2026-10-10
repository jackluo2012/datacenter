package svc

import (
	"datacenter/job/model"
	"datacenter/job/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	SyncTaskModel model.SyncTaskModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Mysql 三无依赖：仅本服务 dc_sync_task 表的 model
	//（DS 客户端由 logic 包按环境变量自装配，不进 svc，避免 logic→svc 反向依赖成环）
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:        c,
		SyncTaskModel: model.NewSyncTaskModel(conn),
	}
}
