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

// 审计日志查询：resource 精确匹配可选 + uid 可选 + 分页
func (l *ListAuditLogsLogic) ListAuditLogs(in *ops.AuditLogListReq) (*ops.AuditLogListReply, error) {
	list, total, err := l.svcCtx.AuditLogModel.FindList(l.ctx, in.Resource, in.Uid, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	logs := make([]*ops.AuditLog, 0, len(list))
	for _, lg := range list {
		logs = append(logs, auditLogToPb(lg))
	}
	return &ops.AuditLogListReply{List: logs, Total: total}, nil
}
