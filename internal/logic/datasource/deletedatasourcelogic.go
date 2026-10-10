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

type DeleteDataSourceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除数据源
func NewDeleteDataSourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDataSourceLogic {
	return &DeleteDataSourceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除数据源: 网关请求逐字段映射为 rpc 请求后转发, 业务错误直接透传
func (l *DeleteDataSourceLogic) DeleteDataSource(req *types.DataSourceIdReq) error {
	_, err := l.svcCtx.DataSourceRpc.DeleteDataSource(l.ctx, &dataSourceClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
