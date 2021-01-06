package config

import "github.com/tal-tech/go-zero/zrpc"

type EserverConfig struct {
	Urls     []string
	User     string
	Password string
}

type Config struct {
	zrpc.RpcServerConf
	Esconfig EserverConfig
}
