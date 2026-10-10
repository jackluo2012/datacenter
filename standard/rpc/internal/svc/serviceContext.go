package svc

import (
	"sync"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// dataSourceEtcdKey datasource rpc 的 etcd 服务 key（Phase 1 契约固定）
const dataSourceEtcdKey = "datasource.rpc"

type ServiceContext struct {
	Config           config.Config
	Conn             sqlx.SqlConn //中台主库连接（dc_std_* 六表）
	WordRootModel    model.WordRootModel
	DictModel        model.DictModel
	DictItemModel    model.DictItemModel
	ElementModel     model.ElementModel
	StdModelModel    model.StdModelModel
	CheckReportModel model.CheckReportModel
	DataSourceRpc    dataSourceClient.DataSource //数据源 rpc（CheckTable 换目标库 DSN）

	connMu     sync.Mutex
	targetCons map[string]sqlx.SqlConn //目标库连接池缓存：key=dsn，复用底层连接池
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	//datasource rpc 客户端：etcd key 固定 datasource.rpc；
	//NonBlock 保证 datasource 未就绪时不阻塞本服务启动，首次调用才建连
	dsRpc := dataSourceClient.NewDataSource(zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: c.Etcd.Hosts,
			Key:   dataSourceEtcdKey,
		},
		NonBlock: true,
	}))
	return &ServiceContext{
		Config:           c,
		Conn:             conn,
		WordRootModel:    model.NewWordRootModel(conn),
		DictModel:        model.NewDictModel(conn),
		DictItemModel:    model.NewDictItemModel(conn),
		ElementModel:     model.NewElementModel(conn),
		StdModelModel:    model.NewStdModelModel(conn),
		CheckReportModel: model.NewCheckReportModel(conn),
		DataSourceRpc:    dsRpc,
		targetCons:       make(map[string]sqlx.SqlConn),
	}
}

// TargetConn 获取目标库连接（按 DSN 缓存 sqlx.SqlConn，复用底层连接池）
func (sc *ServiceContext) TargetConn(dsn string) (sqlx.SqlConn, error) {
	sc.connMu.Lock()
	defer sc.connMu.Unlock()
	if conn, ok := sc.targetCons[dsn]; ok {
		return conn, nil
	}
	conn := sqlx.NewMysql(dsn)
	sc.targetCons[dsn] = conn
	return conn, nil
}
