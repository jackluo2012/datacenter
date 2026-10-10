// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package datasource

import (
	"context"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DataSourceInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 数据源详情
func NewDataSourceInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DataSourceInfoLogic {
	return &DataSourceInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 数据源详情: 网关对外不回传密钥, with_secret 固定为 false
func (l *DataSourceInfoLogic) DataSourceInfo(req *types.DataSourceIdReq) (resp *types.DataSourceInfo, err error) {
	reply, err := l.svcCtx.DataSourceRpc.GetDataSource(l.ctx, &dataSourceClient.GetDataSourceReq{
		Id:         req.Id,
		Beid:       req.Beid,
		WithSecret: false,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return dataSourceInfoToTypes(reply.Datasource), nil
}
