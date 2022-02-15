package logic

import (
	"context"

	"datacenter/questions/model"
	"datacenter/questions/rpc/internal/svc"
	"datacenter/questions/rpc/questions"

	"github.com/zeromicro/go-zero/core/logx"
)

type PostQuestionsChangeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPostQuestionsChangeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostQuestionsChangeLogic {
	return &PostQuestionsChangeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

//  提交 答案
func (l *PostQuestionsChangeLogic) PostQuestionsChange(in *questions.QuestionsAnswerReq) (*questions.QuestionsAnswerResp, error) {
	logx.Info("in=", in)
	result, err := l.svcCtx.QuestionsAnswersModel.Insert(model.AppQuestionsAnswers{
		Beid:       in.Beid,
		Ptyid:      in.Ptyid,
		Answers:    in.Answers,
		Score:      in.Score,
		ActivityId: in.ActivityId,
		Uid:        in.Uid,
		Auid:       in.Auid,
	})
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &questions.QuestionsAnswerResp{
		Score:    in.Score,
		AnswerId: id,
	}, nil
}
