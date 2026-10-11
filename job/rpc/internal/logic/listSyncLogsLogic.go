package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSyncLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSyncLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSyncLogsLogic {
	return &ListSyncLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步日志
func (l *ListSyncLogsLogic) ListSyncLogs(in *job.SyncLogListReq) (*job.SyncLogListReply, error) {
	// todo: add your logic here and delete this line

	return &job.SyncLogListReply{}, nil
}
