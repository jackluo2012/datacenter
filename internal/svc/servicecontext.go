package svc

import (
	"context"
	"datacenter/common/rpc/commonclient"
	"datacenter/internal/config"
	"datacenter/internal/middleware"
	"datacenter/questions/rpc/questionsclient"
	"datacenter/search/rpc/searchclient"
	"datacenter/shared"
	"datacenter/user/rpc/userclient"
	"datacenter/votes/rpc/votesclient"
	"fmt"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config           config.Config
	GreetMiddleware1 rest.Middleware
	GreetMiddleware2 rest.Middleware
	Usercheck        rest.Middleware
	Admincheck       rest.Middleware
	UserRpc          userclient.User           //用户
	CommonRpc        commonclient.Common       //公共
	VotesRpc         votesclient.Votes         //投票
	SearchRpc        searchclient.Search       //搜索
	QuestionsRpc     questionsclient.Questions //问答抽奖
	Cache            cache.Cache
	RedisConn        *redis.Redis
}

func timeInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	stime := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		return err
	}

	fmt.Printf("调用 %s 方法 耗时: %v\n", method, time.Now().Sub(stime))
	return nil
}
func NewServiceContext(c config.Config) *ServiceContext {

	ur := userclient.NewUser(zrpc.MustNewClient(c.UserRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	cr := commonclient.NewCommon(zrpc.MustNewClient(c.CommonRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	vr := votesclient.NewVotes(zrpc.MustNewClient(c.VotesRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	sr := searchclient.NewSearch(zrpc.MustNewClient(c.SearchRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	qr := questionsclient.NewQuestions(zrpc.MustNewClient(c.QuestionsRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	//缓存
	ca := cache.New(c.CacheRedis, syncx.NewSingleFlight(), cache.NewStat("dc"), shared.ErrNotFound)
	rcon := redis.NewRedis(c.CacheRedis[0].Host, c.CacheRedis[0].Type, c.CacheRedis[0].Pass)
	return &ServiceContext{
		Config:           c,
		GreetMiddleware1: greetMiddleware1,
		GreetMiddleware2: greetMiddleware2,
		Usercheck:        middleware.NewUserCheckMiddleware().Handle,
		Admincheck:       middleware.NewAdminCheckMiddleware().Handle,
		UserRpc:          ur,
		CommonRpc:        cr,
		VotesRpc:         vr,
		SearchRpc:        sr,
		QuestionsRpc:     qr,
		Cache:            ca,
		RedisConn:        rcon,
	}
}
func greetMiddleware1(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logx.Info("greetMiddleware1 request ... ")
		next(w, r)
		logx.Info("greetMiddleware1 reponse ... ")
	}
}

func greetMiddleware2(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logx.Info("greetMiddleware2 request ... ")
		next(w, r)
		logx.Info("greetMiddleware2 reponse ... ")
	}
}
