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

type ArchiveLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 归档日志列表
func NewArchiveLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArchiveLogListLogic {
	return &ArchiveLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 归档日志列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *ArchiveLogListLogic) ArchiveLogList(req *types.ArchiveLogListReq) (resp *types.ArchiveLogListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.OpsRpc.ListArchiveLogs(l.ctx, &opsClient.ArchiveLogListReq{
		TaskId: req.TaskId,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.OpsArchiveLog, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *archiveLogToTypes(item))
	}
	return &types.ArchiveLogListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
