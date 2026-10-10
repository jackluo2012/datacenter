// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package datasvc

import (
	"context"

	"datacenter/datasvc/rpc/dataSvcClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新数据API
func NewUpdateApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateApiLogic {
	return &UpdateApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateApi 转发 datasvc rpc 更新数据 API
func (l *UpdateApiLogic) UpdateApi(req *types.UpdateApiReq) error {
	// 逐字段映射为 rpc 请求（网关契约无 resp_sample_json，置空表示不修改时由 rpc 侧判断）
	_, err := l.svcCtx.DataSvcRpc.UpdateApi(l.ctx, &dataSvcClient.UpdateApiReq{
		Id:         req.Id,
		Beid:       req.Beid,
		Name:       req.Name,
		SqlText:    req.SqlText,
		ParamsJson: req.ParamsJson,
		CacheTtl:   req.CacheTtl,
	})
	// rpc 业务错误原样透传
	return err
}
