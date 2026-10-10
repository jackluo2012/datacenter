package logic

import (
	"context"
	"database/sql"
	"time"

	"datacenter/datasource/model"
	"datacenter/datasource/rpc/datasource"
	"datacenter/datasource/rpc/internal/svc"

	_ "github.com/go-sql-driver/mysql" // 注册 mysql 驱动, 供探活使用
	"github.com/zeromicro/go-zero/core/logx"
)

// testConnTimeout 探活超时时间
const testConnTimeout = 3 * time.Second

type TestConnectionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTestConnectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestConnectionLogic {
	return &TestConnectionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 测试连接: mysql 类型真实建连 ping 并记录耗时; 其他类型暂不支持自动探活, 直接返回 ok=true
// 注意: 探活失败不返回 error, 由 TestConnReply.Ok=false + ErrMsg 表达
func (l *TestConnectionLogic) TestConnection(in *datasource.IdReq) (*datasource.TestConnReply, error) {
	// 1.查询数据源(校验存在性与 beid 归属)
	ds, err := l.svcCtx.DataSourceModel.FindOne(in.Id, in.Beid)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, ErrNotFound
		}
		l.Errorf("查询数据源失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, ErrQueryFail
	}
	// 2.非 mysql 类型暂不支持自动探活
	if ds.Type != "mysql" {
		return &datasource.TestConnReply{Ok: true, ErrMsg: "该类型暂不支持自动探活"}, nil
	}
	// 3.解密密码并拼 DSN
	plain, err := DecryptPassword(l.svcCtx.Config.EncryptKey, ds.PasswordEnc)
	if err != nil {
		l.Errorf("解密密码失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, ErrDecryptFail
	}
	dsn := BuildMysqlDsn(ds.Host, ds.Port, ds.DbName, ds.Username, plain)
	// 4.建连 ping, 记录耗时
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return &datasource.TestConnReply{Ok: false, ErrMsg: err.Error()}, nil
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(0)
	ctx, cancel := context.WithTimeout(l.ctx, testConnTimeout)
	defer cancel()
	start := time.Now()
	pingErr := db.PingContext(ctx)
	latencyMs := time.Since(start).Milliseconds()
	if pingErr != nil {
		l.Errorf("数据源探活失败, id: %d, host: %s:%d, err: %v", in.Id, ds.Host, ds.Port, pingErr)
		return &datasource.TestConnReply{Ok: false, LatencyMs: latencyMs, ErrMsg: pingErr.Error()}, nil
	}
	return &datasource.TestConnReply{Ok: true, LatencyMs: latencyMs}, nil
}
