package logic

import (
	"context"

	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunJobLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunJobLogic {
	return &RunJobLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行任务（同步跑完规则并落结果）
func (l *RunJobLogic) RunJob(in *quality.IdReq) (*quality.RunJobReply, error) {
	// todo: add your logic here and delete this line

	return &quality.RunJobReply{}, nil
}
