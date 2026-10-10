package logic

import (
	"context"

	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAppliesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAppliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAppliesLogic {
	return &ListAppliesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 申请列表
func (l *ListAppliesLogic) ListApplies(in *catalog.ApplyListReq) (*catalog.ApplyListReply, error) {
	// todo: add your logic here and delete this line

	return &catalog.ApplyListReply{}, nil
}
