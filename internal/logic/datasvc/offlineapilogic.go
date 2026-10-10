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

type OfflineApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 下线API
func NewOfflineApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OfflineApiLogic {
	return &OfflineApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// OfflineApi 转发 datasvc rpc 下线 API
func (l *OfflineApiLogic) OfflineApi(req *types.DsIdReq) error {
	// 逐字段映射为 rpc 请求
	_, err := l.svcCtx.DataSvcRpc.OfflineApi(l.ctx, &dataSvcClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	// rpc 业务错误原样透传
	return err
}
