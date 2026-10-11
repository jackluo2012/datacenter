package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunSyncLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunSyncLogic {
	return &RunSyncLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行同步（真实跑库：分批读源写目标，同批事务）
func (l *RunSyncLogic) RunSync(in *job.RunSyncReq) (*job.RunSyncReply, error) {
	// todo: add your logic here and delete this line

	return &job.RunSyncReply{}, nil
}
