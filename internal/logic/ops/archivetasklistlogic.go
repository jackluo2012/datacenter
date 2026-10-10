// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package ops

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/ops/rpc/opsClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArchiveTaskListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 归档任务列表
func NewArchiveTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArchiveTaskListLogic {
	return &ArchiveTaskListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 归档任务列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *ArchiveTaskListLogic) ArchiveTaskList(req *types.ArchiveTaskListReq) (resp *types.ArchiveTaskListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.OpsRpc.ListArchiveTasks(l.ctx, &opsClient.TaskListReq{
		Beid: req.Beid,
		Page: page,
		Size: size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.OpsArchiveTask, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *archiveTaskToTypes(item))
	}
	return &types.ArchiveTaskListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
