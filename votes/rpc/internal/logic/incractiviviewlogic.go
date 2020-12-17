package logic

import (
	"context"

	"datacenter/votes/rpc/internal/svc"
	"datacenter/votes/rpc/votes"

	"github.com/tal-tech/go-zero/core/logx"
)

type IncrActiviViewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrActiviViewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrActiviViewLogic {
	return &IncrActiviViewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 增加活动的爆光量
func (l *IncrActiviViewLogic) IncrActiviView(in *votes.ActInfoReq) (*votes.ActInfoResp, error) {
	err := l.svcCtx.AppVotesActivityModel.IncrView(in.Actid)
	return &votes.ActInfoResp{}, err
}
