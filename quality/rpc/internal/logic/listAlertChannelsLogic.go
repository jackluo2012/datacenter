package logic

import (
	"context"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAlertChannelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAlertChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAlertChannelsLogic {
	return &ListAlertChannelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 告警通道分页列表（beid 多租户必带，按 id 倒序）
func (l *ListAlertChannelsLogic) ListAlertChannels(in *quality.AlertChannelListReq) (*quality.AlertChannelListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.AlertChannelModel.FindList(l.ctx, in.Beid, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	channels := make([]*quality.AlertChannel, 0, len(list))
	for _, c := range list {
		channels = append(channels, alertChannelToPb(c))
	}
	return &quality.AlertChannelListReply{List: channels, Total: total}, nil
}
