package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/questions/rpc/questions"

	"github.com/tal-tech/go-zero/core/logx"
)

type ChangeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangeLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChangeLogic {
	return ChangeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangeLogic) Change(in questions.QuestionsAnswerReq) (*questions.QuestionsAnswerResp, error) {

	return l.svcCtx.QuestionsRpc.PostQuestionsChange(l.ctx, &in)

}
