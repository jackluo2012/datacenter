package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"datacenter/internal/config"
	"datacenter/internal/handler"
	"datacenter/internal/middleware"
	"datacenter/internal/svc"
	"datacenter/shared"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/datacenter-api.yaml", "the config file")

func dirhandler(patern, filedir string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		handler := http.StripPrefix(patern, http.FileServer(http.Dir(filedir)))
		handler.ServeHTTP(w, req)

	}
}

func staticFileHandler(engine *rest.Server) {
	//这里注册
	dirlevel := []string{":1", ":2", ":3", ":4", ":5", ":6", ":7", ":8"}
	patern := "/static/"
	dirpath := "./assets/"
	for i := 1; i < len(dirlevel); i++ {
		path := "/" + strings.Join(dirlevel[:i], "/")
		//最后生成 /asset
		engine.AddRoute(
			rest.Route{
				Method:  http.MethodGet,
				Path:    path,
				Handler: dirhandler(patern, dirpath),
			})
	}

}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf, rest.WithNotAllowedHandler(middleware.NewCorsMiddleware().Handler()))
	defer server.Stop()
	server.Use(middleware.NewCorsMiddleware().Handle)
	//静太文件处理
	staticFileHandler(server)
	// 设置错误处理函数
	httpx.SetErrorHandler(shared.ErrorHandler)

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server xxx at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
