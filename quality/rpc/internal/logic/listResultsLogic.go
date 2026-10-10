package logic

import (
	"context"

	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListResultsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListResultsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListResultsLogic {
	return &ListResultsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 结果列表
func (l *ListResultsLogic) ListResults(in *quality.ResultListReq) (*quality.ResultListReply, error) {
	// todo: add your logic here and delete this line

	return &quality.ResultListReply{}, nil
}
