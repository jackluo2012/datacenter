package svc

import (
	"net/http"

	"datacenter/ai/model"
	"datacenter/ai/rpc/internal/config"
	"datacenter/ai/rpc/internal/llm"
	"datacenter/metadata/rpc/metadataClient"
	"datacenter/metric/rpc/metricsvc"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// 跨服务 rpc 的 etcd key（契约固定，与各服务注册的 ListenOn key 一致）
const (
	metadataEtcdKey = "metadata.rpc" //元数据 rpc
	metricEtcdKey   = "metric.rpc"   //指标 rpc
)

type ServiceContext struct {
	Config      config.Config
	Conn        sqlx.SqlConn            //中台主库连接（dc_ask_log 审计表）
	AskLogModel model.AskLogModel       //问数审计 model
	MetadataRpc metadataClient.Metadata //元数据 rpc（语义检索实体）
	MetricRpc   metricsvc.MetricSvc     //指标 rpc（语义检索指标）
	HttpClient  *http.Client            //LLM HTTP 客户端（30s 超时，llm 包内另有 ctx 超时兜底）
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	//config.go 不可改，参考 tag 服务内联装配写法；
	//NonBlock: true 保证 metadata/metric 未就绪时不阻塞本服务启动
	mdRpc := metadataClient.NewMetadata(zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: c.Etcd.Hosts,
			Key:   metadataEtcdKey,
		},
		NonBlock: true,
	}))
	metricRpc := metricsvc.NewMetricSvc(zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: c.Etcd.Hosts,
			Key:   metricEtcdKey,
		},
		NonBlock: true,
	}))
	return &ServiceContext{
		Config:      c,
		Conn:        conn,
		AskLogModel: model.NewAskLogModel(conn),
		MetadataRpc: mdRpc,
		MetricRpc:   metricRpc,
		HttpClient:  &http.Client{Timeout: llm.RequestTimeout},
	}
}
