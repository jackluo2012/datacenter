package logic

import (
	"context"

	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type WriteLoginLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWriteLoginLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WriteLoginLogLogic {
	return &WriteLoginLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 写登录日志（网关登录后调用）
func (l *WriteLoginLogLogic) WriteLoginLog(in *ops.WriteLoginLogReq) (*ops.OkReply, error) {
	// todo: add your logic here and delete this line

	return &ops.OkReply{}, nil
}
