package logic

import (
	"context"

	"datacenter/questions/rpc/internal/svc"
	"datacenter/questions/rpc/questions"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActivitiesInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetActivitiesInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivitiesInfoLogic {
	return &GetActivitiesInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取 问答抽奖活动信息
func (l *GetActivitiesInfoLogic) GetActivitiesInfo(in *questions.ActivitiesReq) (*questions.ActInfoResp, error) {
	// todo: add your logic here and delete this line
	activits, err := l.svcCtx.QuestionsActivitiesModel.FindOne(in.Actid)
	if err != nil {
		return nil, err
	}

	return &questions.ActInfoResp{
		Id:        activits.Id,
		Beid:      activits.Beid,
		Ptyid:     activits.Ptyid,
		StartDate: activits.StartDate,
		EndDate:   activits.EndDate,
		Title:     activits.Title,
		Rule:      activits.Rule,
		GetScore:  int64(activits.GetScore),
		Des:       activits.Des,
		Header:    activits.Header,
		Image:     activits.Image,
	}, nil
}
