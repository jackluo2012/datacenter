package logic

import (
	"context"

	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewApplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewApplyLogic {
	return &ReviewApplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 审批
func (l *ReviewApplyLogic) ReviewApply(in *catalog.ReviewApplyReq) (*catalog.ApplyReply, error) {
	// todo: add your logic here and delete this line

	return &catalog.ApplyReply{}, nil
}
