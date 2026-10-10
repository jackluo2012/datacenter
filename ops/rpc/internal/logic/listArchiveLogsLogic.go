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

// 归档日志（task_id 可选过滤 + 分页，按执行时间倒序）
func (l *ListArchiveLogsLogic) ListArchiveLogs(in *ops.ArchiveLogListReq) (*ops.ArchiveLogListReply, error) {
	list, total, err := l.svcCtx.ArchiveLogModel.FindList(l.ctx, in.TaskId, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	logs := make([]*ops.ArchiveLog, 0, len(list))
	for _, lg := range list {
		logs = append(logs, archiveLogToPb(lg))
	}
	return &ops.ArchiveLogListReply{List: logs, Total: total}, nil
}
