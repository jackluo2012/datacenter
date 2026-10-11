package logic

import (
	"context"
	"strings"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveAlertChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveAlertChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveAlertChannelLogic {
	return &SaveAlertChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 告警通道
// 保存告警通道（id=0 新建默认启用，id>0 更新可编辑字段）：
// 名称必填 + 类型白名单（dingtalk/feishu）+ webhook 必须 http(s) + 触发级别白名单（缺省 failed）
func (l *SaveAlertChannelLogic) SaveAlertChannel(in *quality.SaveAlertChannelReq) (*quality.AlertChannel, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if err := ValidateChannel(in.Name, in.Type, in.WebhookUrl, in.MinLevel); err != nil {
		return nil, err
	}
	// 触发级别缺省回填 failed（仅任务全部不通过才告警）
	minLevel := in.MinLevel
	if minLevel == "" {
		minLevel = model.AlertLevelFailed
	}

	var (
		row *model.AlertChannel
		err error
	)
	if in.Id > 0 {
		// 更新：先确认通道存在（含租户过滤），再更新可编辑字段
		if _, err = l.svcCtx.AlertChannelModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
			return nil, err
		}
		if err = l.svcCtx.AlertChannelModel.Update(l.ctx, &model.AlertChannel{
			Id:         in.Id,
			Name:       strings.TrimSpace(in.Name),
			Type:       in.Type,
			WebhookUrl: strings.TrimSpace(in.WebhookUrl),
			Secret:     in.Secret,
			MinLevel:   minLevel,
			Beid:       in.Beid,
		}); err != nil {
			return nil, err
		}
		row, err = l.svcCtx.AlertChannelModel.FindOne(l.ctx, in.Id, in.Beid)
		if err != nil {
			return nil, err
		}
		return alertChannelToPb(row), nil
	}

	// 新建：默认启用
	id, err := l.svcCtx.AlertChannelModel.Insert(l.ctx, &model.AlertChannel{
		Name:       strings.TrimSpace(in.Name),
		Type:       in.Type,
		WebhookUrl: strings.TrimSpace(in.WebhookUrl),
		Secret:     in.Secret,
		MinLevel:   minLevel,
		Status:     model.ChannelStatusEnabled,
		Beid:       in.Beid,
	})
	if err != nil {
		return nil, err
	}
	row, err = l.svcCtx.AlertChannelModel.FindOne(l.ctx, id, in.Beid)
	if err != nil {
		return nil, err
	}
	return alertChannelToPb(row), nil
}
