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

type UpdateDataSourceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新数据源
func NewUpdateDataSourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDataSourceLogic {
	return &UpdateDataSourceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 更新数据源: 网关请求逐字段映射为 rpc 请求后转发, 业务错误直接透传
func (l *UpdateDataSourceLogic) UpdateDataSource(req *types.UpdateDataSourceReq) error {
	_, err := l.svcCtx.DataSourceRpc.UpdateDataSource(l.ctx, &dataSourceClient.UpdateDataSourceReq{
		Id:         req.Id,
		Beid:       req.Beid,
		Name:       req.Name,
		Type:       req.Type,
		Host:       req.Host,
		Port:       req.Port,
		DbName:     req.DbName,
		Username:   req.Username,
		Password:   req.Password,
		ParamsJson: req.ParamsJson,
		Status:     req.Status,
		Owner:      req.Owner,
	})
	return err
}
