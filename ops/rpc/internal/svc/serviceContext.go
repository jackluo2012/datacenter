package svc

import (
	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/ops/model"
	"datacenter/ops/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config           config.Config
	Conn             sqlx.SqlConn // 中台主库连接（归档任务/归档日志/登录日志/审计日志四表）
	ArchiveTaskModel model.ArchiveTaskModel
	ArchiveLogModel  model.ArchiveLogModel
	LoginLogModel    model.LoginLogModel
	AuditLogModel    model.AuditLogModel
	DataSourceCli    dataSourceClient.DataSource // 数据源服务客户端（RunArchive 取目标库 DSN）
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:           c,
		Conn:             conn,
		ArchiveTaskModel: model.NewArchiveTaskModel(conn),
		ArchiveLogModel:  model.NewArchiveLogModel(conn),
		LoginLogModel:    model.NewLoginLogModel(conn),
		AuditLogModel:    model.NewAuditLogModel(conn),
		// 数据源服务走 etcd 发现（key 固定 datasource.rpc）；NonBlock 启动不阻塞，首次调用才建连
		DataSourceCli: dataSourceClient.NewDataSource(zrpc.MustNewClient(zrpc.RpcClientConf{
			Etcd:     discov.EtcdConf{Hosts: c.Etcd.Hosts, Key: "datasource.rpc"},
			NonBlock: true,
		})),
	}
}
