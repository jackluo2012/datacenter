package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type TurntableLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTurntableLogic(ctx context.Context, svcCtx *svc.ServiceContext) TurntableLogic {
	return TurntableLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TurntableLogic) Turntable(req questionsclient.TurnTableReq) (*questionsclient.AwardInfoResp, error) {

	return l.svcCtx.QuestionsRpc.PostTurnTable(l.ctx, &req)
}
