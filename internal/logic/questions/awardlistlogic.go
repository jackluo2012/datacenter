package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type AwardListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAwardListLogic(ctx context.Context, svcCtx *svc.ServiceContext) AwardListLogic {
	return AwardListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AwardListLogic) AwardList(in questionsclient.ActivitiesReq) (*questionsclient.AwardListResp, error) {

	return l.svcCtx.QuestionsRpc.GetAwardList(l.ctx, &in)

}
