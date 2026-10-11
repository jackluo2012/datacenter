package logic

import (
	"context"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAlertChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAlertChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAlertChannelLogic {
	return &DeleteAlertChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除告警通道（按 id+beid 定位；历史告警日志保留供追溯）
func (l *DeleteAlertChannelLogic) DeleteAlertChannel(in *quality.IdReq) (*quality.IdReq, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrChannelIdRequired
	}
	// 先确认存在（含租户过滤），不存在直接报 NotFound
	if _, err := l.svcCtx.AlertChannelModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	if err := l.svcCtx.AlertChannelModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	return in, nil
}
