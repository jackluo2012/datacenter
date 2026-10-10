package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Mysql struct {
		DataSource string
	}
	EncryptKey string `json:",optional"` //AES-GCM 密钥(32字节)
	CacheRedis cache.ClusterConf `json:",optional"`
}
