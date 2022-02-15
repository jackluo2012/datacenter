package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type LotteryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLotteryLogic(ctx context.Context, svcCtx *svc.ServiceContext) LotteryLogic {
	return LotteryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LotteryLogic) Lottery(in questionsclient.ConvertReq) (*questionsclient.ConvertResp, error) {
	return l.svcCtx.QuestionsRpc.PostConvert(l.ctx, &in)
}
