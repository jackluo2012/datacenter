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

type DataSourceTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 测试数据源连接
func NewDataSourceTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DataSourceTestLogic {
	return &DataSourceTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 测试数据源连接: 网关请求逐字段映射为 rpc 请求后转发 TestConnection, 业务错误直接透传
func (l *DataSourceTestLogic) DataSourceTest(req *types.DataSourceIdReq) (resp *types.TestConnResp, err error) {
	reply, err := l.svcCtx.DataSourceRpc.TestConnection(l.ctx, &dataSourceClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.TestConnResp{
		Ok:        reply.Ok,
		LatencyMs: reply.LatencyMs,
		ErrMsg:    reply.ErrMsg,
	}, nil
}
