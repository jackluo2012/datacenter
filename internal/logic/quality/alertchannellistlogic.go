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

type AlertChannelListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 告警通道列表
func NewAlertChannelListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AlertChannelListLogic {
	return &AlertChannelListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 告警通道列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发;
// 通道列表逐条映射, make 保证空列表序列化为 [] 而非 null
func (l *AlertChannelListLogic) AlertChannelList(req *types.AlertChannelListReq) (resp *types.AlertChannelListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.QualityRpc.ListAlertChannels(l.ctx, &qualityClient.AlertChannelListReq{
		Beid: req.Beid,
		Page: page,
		Size: size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.AlertChannelInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, alertChannelToTypes(item))
	}
	return &types.AlertChannelListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
