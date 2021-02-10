package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type ListsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListsLogic(ctx context.Context, svcCtx *svc.ServiceContext) ListsLogic {
	return ListsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListsLogic) Lists(in questionsclient.ActivitiesReq) (*questionsclient.QuestionsListResp, error) {
	return l.svcCtx.QuestionsRpc.GetQuestionsList(l.ctx, &in)

}
