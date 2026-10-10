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

type CreateDataSourceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建数据源
func NewCreateDataSourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDataSourceLogic {
	return &CreateDataSourceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 创建数据源: 网关请求逐字段映射为 rpc 请求, 转发后把 DataSourceInfo 转为网关类型返回
func (l *CreateDataSourceLogic) CreateDataSource(req *types.CreateDataSourceReq) (resp *types.DataSourceInfo, err error) {
	reply, err := l.svcCtx.DataSourceRpc.CreateDataSource(l.ctx, &dataSourceClient.CreateDataSourceReq{
		Name:       req.Name,
		Type:       req.Type,
		Host:       req.Host,
		Port:       req.Port,
		DbName:     req.DbName,
		Username:   req.Username,
		Password:   req.Password,
		ParamsJson: req.ParamsJson,
		Owner:      req.Owner,
		Beid:       req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return dataSourceInfoToTypes(reply.Datasource), nil
}
