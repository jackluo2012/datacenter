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

type DeleteSyncTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除同步任务
func NewDeleteSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSyncTaskLogic {
	return &DeleteSyncTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除同步任务: 网关请求逐字段映射为 rpc 请求后转发, 业务错误直接透传
func (l *DeleteSyncTaskLogic) DeleteSyncTask(req *types.SyncTaskIdReq) error {
	_, err := l.svcCtx.JobRpc.DeleteSyncTask(l.ctx, &jobClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
