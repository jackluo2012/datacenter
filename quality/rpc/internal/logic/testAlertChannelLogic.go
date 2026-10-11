package logic

import (
	"context"
	"fmt"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestAlertChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTestAlertChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestAlertChannelLogic {
	return &TestAlertChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 测试告警通道：按通道类型发送"【测试】数据质量告警"消息（与真实告警同一发送逻辑：
// 同一消息体构造 + 钉钉加签 + 响应码校验）。
// 发送失败不抛 gRPC 错误，以 TestAlertReply{ok:false,msg} 返回便于前端直接展示
func (l *TestAlertChannelLogic) TestAlertChannel(in *quality.IdReq) (*quality.TestAlertReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrChannelIdRequired
	}
	ch, err := l.svcCtx.AlertChannelModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	content := fmt.Sprintf("【测试】数据质量告警\n通道：%s（%s）\n说明：这是一条连通性测试消息，收到即配置生效", ch.Name, ch.Type)
	if err := sendAlertToChannel(ch, content); err != nil {
		l.Errorf("测试告警通道失败: channel=%d err=%v", ch.Id, err)
		return &quality.TestAlertReply{Ok: false, Msg: err.Error()}, nil
	}
	return &quality.TestAlertReply{Ok: true, Msg: "测试消息已发送"}, nil
}
