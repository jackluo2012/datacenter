package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/votes/rpc/votes"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnrollInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEnrollInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) EnrollInfoLogic {
	return EnrollInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EnrollInfoLogic) EnrollInfo(in *votes.EnrollInfoReq) (*votes.EnrollResp, error) {

	return l.svcCtx.VotesRpc.GetEnrollInfo(l.ctx, in)
}
