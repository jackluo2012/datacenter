package svc

import (
	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/job/model"
	"datacenter/job/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	Conn          sqlx.SqlConn // 中台主库连接（dc_sync_task/dc_sync_task_log + 内置引擎目标库兜底）
	SyncTaskModel model.SyncTaskModel
	SyncLogModel  model.SyncLogModel          // 内置同步执行日志（dc_sync_task_log）
	DataSourceCli dataSourceClient.DataSource // 数据源服务客户端（RunSync 取源/目标库 DSN）
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Mysql 单连接复用：任务两表的 model + 内置引擎目标库兜底（target_ds_id=0 落平台自有库）；
	// DS 客户端与数据源客户端均由 logic/svc 按配置自装配，不进 logic，避免反向依赖成环
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	// dc_sync_task_log 执行日志表启动时自动建（幂等 DDL，新建环境免迁移）；
	// 失败不阻断启动（库暂不可达时只影响执行日志落库/查询，同步主流程照常），记录后放行
	if err := model.EnsureSyncLogTable(conn); err != nil {
		logx.Errorf("dc_sync_task_log 自动建表失败(不影响启动): %v", err)
	}
	return &ServiceContext{
		Config:        c,
		Conn:          conn,
		SyncTaskModel: model.NewSyncTaskModel(conn),
		SyncLogModel:  model.NewSyncLogModel(conn),
		// 数据源服务走 etcd 发现（key 固定 datasource.rpc）；NonBlock 启动不阻塞，首次调用才建连
		DataSourceCli: dataSourceClient.NewDataSource(zrpc.MustNewClient(zrpc.RpcClientConf{
			Etcd:     discov.EtcdConf{Hosts: c.Etcd.Hosts, Key: "datasource.rpc"},
			NonBlock: true,
		})),
	}
}
