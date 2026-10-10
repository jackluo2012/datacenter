package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDistributionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDistributionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDistributionsLogic {
	return &ListDistributionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分发记录
func (l *ListDistributionsLogic) ListDistributions(in *mdm.DistListReq) (*mdm.DistListReply, error) {
	// todo: add your logic here and delete this line

	return &mdm.DistListReply{}, nil
}
