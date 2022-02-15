package config

import "github.com/zeromicro/go-zero/zrpc"

type EserverConfig struct {
	Urls     []string
	User     string
	Password string
}

type Config struct {
	zrpc.RpcServerConf
	Esconfig EserverConfig
}
