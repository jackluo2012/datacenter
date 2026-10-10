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

type DataSourceListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 数据源列表
func NewDataSourceListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DataSourceListLogic {
	return &DataSourceListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 数据源列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *DataSourceListLogic) DataSourceList(req *types.DataSourceListReq) (resp *types.DataSourceListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.DataSourceRpc.ListDataSources(l.ctx, &dataSourceClient.ListReq{
		Beid: req.Beid,
		Type: req.Type,
		Page: page,
		Size: size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.DataSourceInfo, 0, len(reply.List))
	for _, item := range reply.List {
		list = append(list, *dataSourceInfoToTypes(item))
	}
	return &types.DataSourceListResp{
		List:  list,
		Total: reply.Total,
	}, nil
}
