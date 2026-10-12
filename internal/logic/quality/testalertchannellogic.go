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

type TestAlertChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 测试告警通道
func NewTestAlertChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestAlertChannelLogic {
	return &TestAlertChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 测试告警通道: 按 id/beid 触发 rpc 侧真实推送, 推送结果(ok/msg)映射为网关类型返回;
// 发送失败不算业务错误, 以 ok=false 透出
func (l *TestAlertChannelLogic) TestAlertChannel(req *types.QualityIdReq) (resp *types.TestAlertResp, err error) {
	reply, err := l.svcCtx.QualityRpc.TestAlertChannel(l.ctx, &qualityClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.TestAlertResp{
		Ok:  reply.GetOk(),
		Msg: reply.GetMsg(),
	}, nil
}
