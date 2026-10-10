package logic

import (
	"context"

	"datacenter/ops/model"
	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type WriteAuditLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWriteAuditLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WriteAuditLogLogic {
	return &WriteAuditLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 写通用审计日志：直接 Insert。与登录日志同策略——写接口不向上返回错误，
// 审计失败不能影响业务主流程，失败仅记服务端日志，通过 OkReply.Ok=false 弱提示
func (l *WriteAuditLogLogic) WriteAuditLog(in *ops.WriteAuditLogReq) (*ops.OkReply, error) {
	entry := &model.AuditLog{
		Uid:        in.Uid,
		Username:   truncateRunes(in.Username, usernameMaxLen),
		Action:     truncateRunes(in.Action, actionMaxLen),
		Resource:   truncateRunes(in.Resource, resourceMaxLen),
		ResourceId: truncateRunes(in.ResourceId, resourceIdMaxLen),
		DetailJson: truncateBytes(in.DetailJson, detailJsonMaxByte),
		Ip:         truncateRunes(in.Ip, ipMaxLen),
	}
	if err := l.svcCtx.AuditLogModel.Insert(l.ctx, entry); err != nil {
		l.Errorf("写审计日志失败 uid=%d action=%s resource=%s: %v", in.Uid, in.Action, in.Resource, err)
		return &ops.OkReply{Ok: false}, nil
	}
	return &ops.OkReply{Ok: true}, nil
}
