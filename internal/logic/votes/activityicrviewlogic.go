package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/votes/rpc/votes"

	"github.com/zeromicro/go-zero/core/logx"
)

type ActivityIcrViewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivityIcrViewLogic(ctx context.Context, svcCtx *svc.ServiceContext) ActivityIcrViewLogic {
	return ActivityIcrViewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivityIcrViewLogic) ActivityIcrView(req votes.ActInfoReq) (*votes.ActInfoResp, error) {
	return l.svcCtx.VotesRpc.IncrActiviView(l.ctx, &req)
}
