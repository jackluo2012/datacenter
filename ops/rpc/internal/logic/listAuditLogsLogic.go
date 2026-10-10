package logic

import (
	"context"

	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAuditLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAuditLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAuditLogsLogic {
	return &ListAuditLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 审计日志查询
func (l *ListAuditLogsLogic) ListAuditLogs(in *ops.AuditLogListReq) (*ops.AuditLogListReply, error) {
	// todo: add your logic here and delete this line

	return &ops.AuditLogListReply{}, nil
}
