package logic

import (
	"context"

	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteApiLogic {
	return &ExecuteApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行 API（核心：鉴权→限流→缓存→执行→脱敏）
func (l *ExecuteApiLogic) ExecuteApi(in *datasvc.ExecuteReq) (*datasvc.ExecuteReply, error) {
	// todo: add your logic here and delete this line

	return &datasvc.ExecuteReply{}, nil
}
