package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	UserRpc      zrpc.RpcClientConf
	CommonRpc    zrpc.RpcClientConf
	VotesRpc     zrpc.RpcClientConf
	SearchRpc    zrpc.RpcClientConf
	QuestionsRpc zrpc.RpcClientConf

	//Phase 1 数据中台服务
	DataSourceRpc zrpc.RpcClientConf
	MetadataRpc   zrpc.RpcClientConf
	DataSvcRpc    zrpc.RpcClientConf
	CatalogRpc    zrpc.RpcClientConf
	QualityRpc    zrpc.RpcClientConf
	MetricRpc     zrpc.RpcClientConf
	TagRpc        zrpc.RpcClientConf

	//Phase 3/4 数据中台服务
	JobRpc zrpc.RpcClientConf
	IdeRpc zrpc.RpcClientConf
	AiRpc  zrpc.RpcClientConf

	//Phase 5 数据中台服务
	StandardRpc zrpc.RpcClientConf
	MdmRpc      zrpc.RpcClientConf
	OpsRpc      zrpc.RpcClientConf
	ThemeRpc    zrpc.RpcClientConf
	SecurityRpc zrpc.RpcClientConf

	//Phase 7 系统管理 RBAC
	SysRpc zrpc.RpcClientConf

	CacheRedis cache.ClusterConf
}
