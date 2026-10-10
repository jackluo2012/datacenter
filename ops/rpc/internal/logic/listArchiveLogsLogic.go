package logic

import (
	"context"

	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListArchiveLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListArchiveLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListArchiveLogsLogic {
	return &ListArchiveLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 归档日志
func (l *ListArchiveLogsLogic) ListArchiveLogs(in *ops.ArchiveLogListReq) (*ops.ArchiveLogListReply, error) {
	// todo: add your logic here and delete this line

	return &ops.ArchiveLogListReply{}, nil
}
