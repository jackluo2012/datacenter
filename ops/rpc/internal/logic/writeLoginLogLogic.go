package logic

import (
	"context"

	"datacenter/ops/model"
	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type WriteLoginLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWriteLoginLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WriteLoginLogLogic {
	return &WriteLoginLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 写登录日志（网关登录后调用）：直接 Insert（username/ip/ua/success/msg）。
// 写接口不向上返回错误——登录主流程不能被日志写入拖垮，失败仅记服务端日志，
// 通过 OkReply.Ok=false 向调用方弱提示
func (l *WriteLoginLogLogic) WriteLoginLog(in *ops.WriteLoginLogReq) (*ops.OkReply, error) {
	entry := &model.LoginLog{
		Username:  truncateRunes(in.Username, usernameMaxLen),
		Uid:       in.Uid,
		Ip:        truncateRunes(in.Ip, ipMaxLen),
		UserAgent: truncateRunes(in.UserAgent, userAgentMaxLen),
		Success:   b2i(in.Success),
		Msg:       truncateRunes(in.Msg, logMsgMaxLen),
	}
	if err := l.svcCtx.LoginLogModel.Insert(l.ctx, entry); err != nil {
		l.Errorf("写登录日志失败 username=%s ip=%s: %v", in.Username, in.Ip, err)
		return &ops.OkReply{Ok: false}, nil
	}
	return &ops.OkReply{Ok: true}, nil
}
