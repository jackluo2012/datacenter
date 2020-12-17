package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/votes/rpc/votes"

	"github.com/tal-tech/go-zero/core/logx"
)

type ActivityInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivityInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) ActivityInfoLogic {
	return ActivityInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivityInfoLogic) ActivityInfo(req votes.ActInfoReq) (*votes.ActInfoResp, error) {

	return l.svcCtx.VotesRpc.GetActivityInfo(l.ctx, &req)
}
