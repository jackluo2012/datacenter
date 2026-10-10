package svc

import (
	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	Conn          sqlx.SqlConn // 中台主库连接（质量规则/任务/结果三表）
	RuleModel     model.RuleModel
	JobModel      model.JobModel
	ResultModel   model.ResultModel
	DataSourceCli dataSourceClient.DataSource // 数据源服务客户端（RunJob 取目标库 DSN）
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:      c,
		Conn:        conn,
		RuleModel:   model.NewRuleModel(conn),
		JobModel:    model.NewJobModel(conn),
		ResultModel: model.NewResultModel(conn),
		// 数据源服务走 etcd 发现（key 固定 datasource.rpc）；NonBlock 启动不阻塞，首次调用才建连
		DataSourceCli: dataSourceClient.NewDataSource(zrpc.MustNewClient(zrpc.RpcClientConf{
			Etcd:     discov.EtcdConf{Hosts: c.Etcd.Hosts, Key: "datasource.rpc"},
			NonBlock: true,
		})),
	}
}
