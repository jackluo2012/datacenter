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

// 同步日志（dc_sync_task_log 分页倒序，task_id>0 时按任务过滤）
func (l *ListSyncLogsLogic) ListSyncLogs(in *job.SyncLogListReq) (*job.SyncLogListReply, error) {
	list, total, err := l.svcCtx.SyncLogModel.FindList(l.ctx, in.TaskId, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	logs := make([]*job.SyncLog, 0, len(list))
	for _, r := range list {
		logs = append(logs, &job.SyncLog{
			Id:        r.Id,
			TaskId:    r.TaskId,
			ReadRows:  r.ReadRows,
			WriteRows: r.WriteRows,
			Status:    r.Status,
			Msg:       r.Msg,
			RunAt:     r.RunAt,
		})
	}
	return &job.SyncLogListReply{List: logs, Total: total}, nil
}
