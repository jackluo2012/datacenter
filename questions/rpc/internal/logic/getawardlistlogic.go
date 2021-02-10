package logic

import (
	"context"

	"datacenter/questions/rpc/internal/svc"
	"datacenter/questions/rpc/questions"

	"github.com/tal-tech/go-zero/core/logx"
)

type GetAwardListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAwardListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAwardListLogic {
	return &GetAwardListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取 问答奖品列表
func (l *GetAwardListLogic) GetAwardList(in *questions.ActivitiesReq) (*questions.AwardListResp, error) {
	lists, err := l.svcCtx.QuestionsAwardsModel.Find(in.Actid)
	if err != nil {
		return nil, err
	}
	list := make([]*questions.AwardInfoResp, 0)
	if len(lists) > 0 {
		for _, award := range lists {
			list = append(list, Awards2AwardInfoResp(&award))
		}
	}
	return &questions.AwardListResp{
		Data: list,
	}, nil
}
