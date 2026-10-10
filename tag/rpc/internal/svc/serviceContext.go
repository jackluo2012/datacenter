package svc

import (
	"sync"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// dataSourceEtcdKey datasource rpc 的 etcd 服务 key（Phase 1 契约固定）
const dataSourceEtcdKey = "datasource.rpc"

type ServiceContext struct {
	Config        config.Config
	Conn          sqlx.SqlConn //中台主库连接（dc_tag_* 三表）
	ObjectModel   model.ObjectModel
	TagDefModel   model.TagDefModel
	TagValueModel model.TagValueModel
	DataSourceRpc dataSourceClient.DataSource //数据源 rpc（Tagging 换目标库 DSN）

	connMu     sync.Mutex
	targetCons map[string]sqlx.SqlConn //目标库连接池缓存：key=dsn，复用底层连接池
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	//datasource rpc 客户端：config.go 不可改，参考 datasvc 内联装配写法，
	//etcd key 固定 datasource.rpc；NonBlock 保证 datasource 未就绪时不阻塞本服务启动
	dsRpc := dataSourceClient.NewDataSource(zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: c.Etcd.Hosts,
			Key:   dataSourceEtcdKey,
		},
		NonBlock: true,
	}))
	return &ServiceContext{
		Config:        c,
		Conn:          conn,
		ObjectModel:   model.NewObjectModel(conn),
		TagDefModel:   model.NewTagDefModel(conn),
		TagValueModel: model.NewTagValueModel(conn),
		DataSourceRpc: dsRpc,
		targetCons:    make(map[string]sqlx.SqlConn),
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
