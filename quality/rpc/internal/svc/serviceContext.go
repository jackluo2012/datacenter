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
	Config            config.Config
	Conn              sqlx.SqlConn // 中台主库连接（质量规则/任务/结果三表）
	RuleModel         model.RuleModel
	JobModel          model.JobModel
	ResultModel       model.ResultModel
	AlertChannelModel model.AlertChannelModel     // 告警通道（通道管理 + RunJob 告警钩子）
	AlertLogModel     model.AlertLogModel         // 告警发送日志
	ErrorDataModel    model.ErrorDataModel        // 错误数据样本（失败规则下钻）
	ProfileModel      model.ProfileModel          // 表剖析与健康分
	StatsModel        model.StatsModel            // 质量统计聚合（质量大盘：通过率/失败 TopN/健康分趋势）
	DataSourceCli     dataSourceClient.DataSource // 数据源服务客户端（RunJob/ProfileTable 取目标库 DSN）
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:            c,
		Conn:              conn,
		RuleModel:         model.NewRuleModel(conn),
		JobModel:          model.NewJobModel(conn),
		ResultModel:       model.NewResultModel(conn),
		AlertChannelModel: model.NewAlertChannelModel(conn),
		AlertLogModel:     model.NewAlertLogModel(conn),
		ErrorDataModel:    model.NewErrorDataModel(conn),
		ProfileModel:      model.NewProfileModel(conn),
		StatsModel:        model.NewStatsModel(conn),
		// 数据源服务走 etcd 发现（key 固定 datasource.rpc）；NonBlock 启动不阻塞，首次调用才建连
		DataSourceCli: dataSourceClient.NewDataSource(zrpc.MustNewClient(zrpc.RpcClientConf{
			Etcd:     discov.EtcdConf{Hosts: c.Etcd.Hosts, Key: "datasource.rpc"},
			NonBlock: true,
		})),
	}
}
