package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Mysql struct {
		DataSource string
	}
	Llm struct {
		Enabled bool   `json:",default=false"`
		BaseUrl string `json:",optional"` //OpenAI 兼容接口,如 https://api.openai.com/v1
		ApiKey  string `json:",optional"`
		Model   string `json:",default=gpt-4o-mini"`
	}
}
