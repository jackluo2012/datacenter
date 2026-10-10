// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package job

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/job/rpc/jobClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncTaskListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 同步任务列表
func NewSyncTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncTaskListLogic {
	return &SyncTaskListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 同步任务列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *SyncTaskListLogic) SyncTaskList(req *types.SyncTaskListReq) (resp *types.SyncTaskListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.JobRpc.ListSyncTasks(l.ctx, &jobClient.TaskListReq{
		Beid:   req.Beid,
		Status: req.Status,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.SyncTask, 0, len(reply.List))
	for _, item := range reply.List {
		list = append(list, *syncTaskToTypes(item))
	}
	return &types.SyncTaskListResp{
		List:  list,
		Total: reply.Total,
	}, nil
}
