// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package metadata

import (
	"context"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/metadata/rpc/metadataClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CollectDatabaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 采集库表元数据
func NewCollectDatabaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectDatabaseLogic {
	return &CollectDatabaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 采集库表元数据: dsn 为空且指定了数据源时, 先经数据源服务(with_secret=true)换取 dsn, 再转发元数据采集
func (l *CollectDatabaseLogic) CollectDatabase(req *types.CollectDatabaseReq) (resp *types.CollectResp, err error) {
	dsn := req.Dsn
	if dsn == "" && req.DsId > 0 {
		// 内部服务调用, 携带 with_secret 解密密码并拼 dsn
		dsReply, err := l.svcCtx.DataSourceRpc.GetDataSource(l.ctx, &dataSourceClient.GetDataSourceReq{
			Id:         req.DsId,
			Beid:       req.Beid,
			WithSecret: true,
		})
		if err != nil {
			// 业务错误直接透传 rpc error
			return nil, err
		}
		dsn = dsReply.Dsn
	}
	// db_type 缺省按 mysql 处理
	dbType := req.DbType
	if dbType == "" {
		dbType = "mysql"
	}
	reply, err := l.svcCtx.MetadataRpc.CollectDatabase(l.ctx, &metadataClient.CollectDatabaseReq{
		DsId:   req.DsId,
		Beid:   req.Beid,
		DbType: dbType,
		Dsn:    dsn,
		DbName: req.DbName,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.CollectResp{
		TablesCreated:  reply.TablesCreated,
		ColumnsCreated: reply.ColumnsCreated,
	}, nil
}
