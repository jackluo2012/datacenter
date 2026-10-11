package svc

import (
	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	Conn      sqlx.SqlConn //中台主库连接（sys_user/sys_role/sys_menu 等五表）
	UserModel model.UserModel
	RoleModel model.RoleModel
	MenuModel model.MenuModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:    c,
		Conn:      conn,
		UserModel: model.NewUserModel(conn),
		RoleModel: model.NewRoleModel(conn),
		MenuModel: model.NewMenuModel(conn),
	}
}
