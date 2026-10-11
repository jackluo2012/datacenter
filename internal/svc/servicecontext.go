package svc

import (
	"context"
	"datacenter/ai/rpc/aiClient"
	"datacenter/catalog/rpc/catalogClient"
	"datacenter/common/rpc/commonclient"
	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/datasvc/rpc/dataSvcClient"
	"datacenter/ide/rpc/ideClient"
	"datacenter/internal/config"
	"datacenter/internal/middleware"
	"datacenter/job/rpc/jobClient"
	"datacenter/mdm/rpc/mdmClient"
	"datacenter/metadata/rpc/metadataClient"
	"datacenter/metric/rpc/metricsvc"
	"datacenter/ops/rpc/opsClient"
	"datacenter/quality/rpc/qualityClient"
	"datacenter/questions/rpc/questionsclient"
	"datacenter/search/rpc/searchclient"
	"datacenter/security/rpc/securityClient"
	"datacenter/shared"
	"datacenter/standard/rpc/standardClient"
	"datacenter/sys/rpc/sysClient"
	"datacenter/tag/rpc/tagClient"
	"datacenter/theme/rpc/themeClient"
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
	AuditLog         rest.Middleware             //审计切面（管理面写操作自动落审计日志）
	UserRpc          userclient.User             //用户
	CommonRpc        commonclient.Common         //公共
	VotesRpc         votesclient.Votes           //投票
	SearchRpc        searchclient.Search         //搜索
	QuestionsRpc     questionsclient.Questions   //问答抽奖
	DataSourceRpc    dataSourceClient.DataSource //数据源（Phase 1）
	MetadataRpc      metadataClient.Metadata     //元数据（Phase 1）
	DataSvcRpc       dataSvcClient.DataSvc       //数据服务（Phase 1）
	CatalogRpc       catalogClient.Catalog       //资产目录（Phase 1）
	QualityRpc       qualityClient.Quality       //数据质量（Phase 2）
	MetricRpc        metricsvc.MetricSvc         //数据指标（Phase 2）
	TagRpc           tagClient.Tag               //数据标签（Phase 2）
	JobRpc           jobClient.Job               //调度集成（Phase 3）
	IdeRpc           ideClient.Ide               //SQL IDE（Phase 3）
	AiRpc            aiClient.Ai                 //AI 问数（Phase 4）
	StandardRpc      standardClient.Standard     //数据标准（Phase 5）
	MdmRpc           mdmClient.Mdm               //主数据（Phase 5）
	OpsRpc           opsClient.Ops               //数据运维与日志（Phase 5）
	ThemeRpc         themeClient.Theme           //数仓建模（Phase 6）
	SecurityRpc      securityClient.Security     //数据安全（Phase 6）
	SysRpc           sysClient.Sys               //系统管理 RBAC（Phase 7）
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
	dsr := dataSourceClient.NewDataSource(zrpc.MustNewClient(c.DataSourceRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	mr := metadataClient.NewMetadata(zrpc.MustNewClient(c.MetadataRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	dvr := dataSvcClient.NewDataSvc(zrpc.MustNewClient(c.DataSvcRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	cr2 := catalogClient.NewCatalog(zrpc.MustNewClient(c.CatalogRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	qr2 := qualityClient.NewQuality(zrpc.MustNewClient(c.QualityRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	mr2 := metricsvc.NewMetricSvc(zrpc.MustNewClient(c.MetricRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	tr2 := tagClient.NewTag(zrpc.MustNewClient(c.TagRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	jr3 := jobClient.NewJob(zrpc.MustNewClient(c.JobRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	ir3 := ideClient.NewIde(zrpc.MustNewClient(c.IdeRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	ar3 := aiClient.NewAi(zrpc.MustNewClient(c.AiRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	sr5 := standardClient.NewStandard(zrpc.MustNewClient(c.StandardRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	mr5 := mdmClient.NewMdm(zrpc.MustNewClient(c.MdmRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	or5 := opsClient.NewOps(zrpc.MustNewClient(c.OpsRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	tr6 := themeClient.NewTheme(zrpc.MustNewClient(c.ThemeRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	sr6 := securityClient.NewSecurity(zrpc.MustNewClient(c.SecurityRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	scr7 := sysClient.NewSys(zrpc.MustNewClient(c.SysRpc, zrpc.WithUnaryClientInterceptor(timeInterceptor)))
	//缓存
	ca := cache.New(c.CacheRedis, syncx.NewSingleFlight(), cache.NewStat("dc"), shared.ErrNotFound)
	rcon := redis.MustNewRedis(c.CacheRedis[0].RedisConf)
	return &ServiceContext{
		Config:           c,
		GreetMiddleware1: greetMiddleware1,
		GreetMiddleware2: greetMiddleware2,
		Usercheck:        middleware.NewUserCheckMiddleware().Handle,
		Admincheck:       middleware.NewAdminCheckMiddleware().Handle,
		AuditLog:         middleware.NewAuditLogMiddleware(or5).Handle, //审计切面挂载到全局 Use
		UserRpc:          ur,
		CommonRpc:        cr,
		VotesRpc:         vr,
		SearchRpc:        sr,
		QuestionsRpc:     qr,
		DataSourceRpc:    dsr,
		MetadataRpc:      mr,
		DataSvcRpc:       dvr,
		CatalogRpc:       cr2,
		QualityRpc:       qr2,
		MetricRpc:        mr2,
		TagRpc:           tr2,
		JobRpc:           jr3,
		IdeRpc:           ir3,
		AiRpc:            ar3,
		StandardRpc:      sr5,
		MdmRpc:           mr5,
		OpsRpc:           or5,
		ThemeRpc:         tr6,
		SecurityRpc:      sr6,
		SysRpc:           scr7,
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
