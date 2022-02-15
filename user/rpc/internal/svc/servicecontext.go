package svc

import (
	"datacenter/common/rpc/commonclient"
	"datacenter/user/model"
	"datacenter/user/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	c            config.Config
	UserModel    *model.BaseMemberModel
	AppUserModel *model.AppUserModel
	CommonRpc    commonclient.Common
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	um := model.NewBaseMemberModel(conn, c.CacheRedis)
	aup := model.NewAppUserModel(conn, c.CacheRedis)
	//调用核心的公共配置 后面再来做难
	comm := commonclient.NewCommon(zrpc.MustNewClient(c.CommonRpc))
	//userclient.NewUser(zrpc.MustNewClient(c.UserRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))

	return &ServiceContext{
		c:            c,
		UserModel:    um,
		CommonRpc:    comm,
		AppUserModel: aup,
	}
}
