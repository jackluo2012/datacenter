package svc

import (
	"datacenter/security/model"
	"datacenter/security/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	Conn           sqlx.SqlConn //中台主库连接（dc_sec_* 四表）
	LevelModel     model.LevelModel
	EntityTagModel model.EntityTagModel
	MaskRuleModel  model.MaskRuleModel
	GrantModel     model.GrantModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:         c,
		Conn:           conn,
		LevelModel:     model.NewLevelModel(conn),
		EntityTagModel: model.NewEntityTagModel(conn),
		MaskRuleModel:  model.NewMaskRuleModel(conn),
		GrantModel:     model.NewGrantModel(conn),
	}
}
