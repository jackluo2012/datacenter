package logic

import (
	"context"

	"datacenter/ops/model"
	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteArchiveTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteArchiveTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteArchiveTaskLogic {
	return &DeleteArchiveTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除任务（按 id+beid 定位；历史执行日志不级联清理，供问题追溯）
func (l *DeleteArchiveTaskLogic) DeleteArchiveTask(in *ops.IdReq) (*ops.IdReq, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, model.ErrTaskIdRequired
	}
	if err := l.svcCtx.ArchiveTaskModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	return &ops.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
