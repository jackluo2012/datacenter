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

type DeleteArchiveTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除归档任务
func NewDeleteArchiveTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteArchiveTaskLogic {
	return &DeleteArchiveTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除归档任务: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteArchiveTask 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteArchiveTaskLogic) DeleteArchiveTask(req *types.OpsIdReq) error {
	_, err := l.svcCtx.OpsRpc.DeleteArchiveTask(l.ctx, &opsClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
