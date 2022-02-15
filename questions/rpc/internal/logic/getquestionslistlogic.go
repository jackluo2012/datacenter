package logic

import (
	"context"

	"datacenter/questions/model"
	"datacenter/questions/rpc/internal/svc"
	"datacenter/questions/rpc/questions"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQuestionsListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func Questions2QuestionsInfoResp(q *model.AppQuestions) *questions.QuestionsResp {
	return &questions.QuestionsResp{
		Id:         q.Id,
		Beid:       q.Beid,
		Ptyid:      q.Ptyid,
		ActivityId: q.ActivityId,
		Options:    q.Options,
		Corrent:    q.Corrent,
		Question:   q.Question,
		Status:     q.Status,
	}
}

func NewGetQuestionsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQuestionsListLogic {
	return &GetQuestionsListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取 问题列表
func (l *GetQuestionsListLogic) GetQuestionsList(in *questions.ActivitiesReq) (*questions.QuestionsListResp, error) {
	lists, err := l.svcCtx.QuestionsModel.Find(in.Actid)
	if err != nil {
		return nil, err
	}
	list := make([]*questions.QuestionsResp, 0)
	if len(lists) > 0 {
		for _, questions := range lists {
			list = append(list, Questions2QuestionsInfoResp(&questions))
		}
	}
	return &questions.QuestionsListResp{
		Data: list,
	}, nil
}
