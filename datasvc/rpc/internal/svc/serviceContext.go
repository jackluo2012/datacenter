package svc

import (
	"errors"
	"sync"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/zrpc"
	"golang.org/x/time/rate"
)

// ErrCacheNotFound 结果缓存未命中（cache.Cache.Get 未命中时返回）
var ErrCacheNotFound = errors.New("cache not found")

// dataSourceEtcdKey datasource rpc 的 etcd 服务 key（PHASE1 契约固定）
const dataSourceEtcdKey = "datasource.rpc"

type ServiceContext struct {
	Config        config.Config
	ApiModel      model.ApiModel
	AppModel      model.AppModel
	GrantModel    model.GrantModel
	CallLogModel  model.CallLogModel          //API 调用日志（Phase 7 可观测埋点）
	ResultCache   cache.Cache                 //API 结果缓存（CacheRedis）
	DataSourceRpc dataSourceClient.DataSource //数据源 rpc（ExecuteApi 获取目标库 DSN）

	limiterMu sync.Mutex
	limiters  map[int64]*rate.Limiter //应用限流器容器：key=app_id，进程内令牌桶

	connMu     sync.Mutex
	targetCons map[string]sqlx.SqlConn //目标库连接池缓存：key=dsn，复用 *sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	//datasource rpc 客户端：config.go 不可改，参考网关 servicecontext 写法内联装配，
	//etcd key 固定为 datasource.rpc；NonBlock 保证 datasource 未就绪时不阻塞本服务启动
	dsRpc := dataSourceClient.NewDataSource(zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: c.Etcd.Hosts,
			Key:   dataSourceEtcdKey,
		},
		NonBlock: true,
	}))
	return &ServiceContext{
		Config:        c,
		ApiModel:      model.NewApiModel(conn),
		AppModel:      model.NewAppModel(conn),
		GrantModel:    model.NewGrantModel(conn),
		CallLogModel:  model.NewCallLogModel(conn),
		ResultCache:   cache.New(c.CacheRedis, syncx.NewSingleFlight(), cache.NewStat("dsvc"), ErrCacheNotFound),
		DataSourceRpc: dsRpc,
		limiters:      make(map[int64]*rate.Limiter),
		targetCons:    make(map[string]sqlx.SqlConn),
	}
}

// Allow 应用级限流：每应用一个进程内令牌桶（map+锁保护），
// 限流配置变更时重建；qps<=0 视为不限流
func (sc *ServiceContext) Allow(appId int64, qps int64) bool {
	if qps <= 0 {
		return true
	}
	sc.limiterMu.Lock()
	lim, ok := sc.limiters[appId]
	if !ok || int64(lim.Limit()) != qps {
		lim = rate.NewLimiter(rate.Limit(qps), int(qps)) //桶容量=qps，允许 1 秒内小突发
		sc.limiters[appId] = lim
	}
	sc.limiterMu.Unlock()
	return lim.Allow()
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
