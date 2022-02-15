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

	CacheRedis cache.ClusterConf
}
