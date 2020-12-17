package main

import (
	"flag"
	"fmt"
	"net/http"

	"datacenter/internal/config"
	"datacenter/internal/handler"
	"datacenter/internal/svc"
	"datacenter/shared"

	"github.com/tal-tech/go-zero/core/conf"
	"github.com/tal-tech/go-zero/rest"
	"github.com/tal-tech/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/datacenter-api.yaml", "the config file")

func corsHandle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS,DELETE,PUT")
		w.Header().Set("Access-Control-Allow-Headers", "content-type,authorization,cookies")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next(w, r)
	}
}

func dirhandler(patern, filedir string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		handler := http.StripPrefix(patern, http.FileServer(http.Dir(filedir)))
		handler.ServeHTTP(w, req)

	}
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()
	server.Use(corsHandle)

	// 设置错误处理函数
	httpx.SetErrorHandler(shared.ErrorHandler)

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server xxx at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
