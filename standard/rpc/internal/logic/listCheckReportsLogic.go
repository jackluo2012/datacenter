package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCheckReportsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCheckReportsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCheckReportsLogic {
	return &ListCheckReportsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 核对历史
func (l *ListCheckReportsLogic) ListCheckReports(in *standard.CheckListReq) (*standard.CheckListReply, error) {
	// todo: add your logic here and delete this line

	return &standard.CheckListReply{}, nil
}
