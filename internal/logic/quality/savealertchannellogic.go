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

type SaveAlertChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存告警通道
func NewSaveAlertChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveAlertChannelLogic {
	return &SaveAlertChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存告警通道: 网关请求逐字段映射为 rpc 请求(含 Secret 密钥字段)转发,
// 转发后把落库后的通道信息(含 rpc 生成的 Id/Status)转为网关类型返回
func (l *SaveAlertChannelLogic) SaveAlertChannel(req *types.SaveAlertChannelReq) (resp *types.AlertChannelInfo, err error) {
	reply, err := l.svcCtx.QualityRpc.SaveAlertChannel(l.ctx, &qualityClient.SaveAlertChannelReq{
		Id:         0, // 网关创建契约无 id 字段, 固定走新建语义, 由 rpc 侧生成
		Name:       req.Name,
		Type:       req.Type,
		WebhookUrl: req.WebhookUrl,
		Secret:     req.Secret,
		MinLevel:   req.MinLevel,
		Beid:       req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	info := alertChannelToTypes(reply)
	return &info, nil
}
