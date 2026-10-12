// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package quality

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/quality/rpc/qualityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAlertChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除告警通道
func NewDeleteAlertChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAlertChannelLogic {
	return &DeleteAlertChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除告警通道: 网关 id/beid 逐字段映射为 rpc IdReq 转发, 业务错误直接透传
func (l *DeleteAlertChannelLogic) DeleteAlertChannel(req *types.QualityIdReq) error {
	_, err := l.svcCtx.QualityRpc.DeleteAlertChannel(l.ctx, &qualityClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return err
	}
	return nil
}
