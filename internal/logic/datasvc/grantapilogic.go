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

type GrantApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 授权API给应用
func NewGrantApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrantApiLogic {
	return &GrantApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GrantApi 转发 datasvc rpc 授权 API 给应用
func (l *GrantApiLogic) GrantApi(req *types.GrantApiReq) error {
	// 逐字段映射为 rpc 请求；expire_at 为 unix 秒，
	// proto 约定 0 表示永久授权，故网关未传时直接透传 0 即可
	_, err := l.svcCtx.DataSvcRpc.GrantApi(l.ctx, &dataSvcClient.GrantApiReq{
		ApiId:     req.ApiId,
		AppId:     req.AppId,
		ExpireAt:  req.ExpireAt,
		FieldMask: req.FieldMask,
		Beid:      req.Beid,
	})
	// rpc 业务错误原样透传
	return err
}
