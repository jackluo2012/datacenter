package logic

import (
	"context"

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

// 写通用审计日志
func (l *WriteAuditLogLogic) WriteAuditLog(in *ops.WriteAuditLogReq) (*ops.OkReply, error) {
	// todo: add your logic here and delete this line

	return &ops.OkReply{}, nil
}
