package logic

import (
	"context"

	"datacenter/questions/model"
	"datacenter/questions/rpc/internal/svc"
	"datacenter/questions/rpc/questions"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAwardInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

/**
 * 数据格式转换
 */
func Awards2AwardInfoResp(awards *model.AppQuestionsAwards) *questions.AwardInfoResp {
	return &questions.AwardInfoResp{
		Id:               awards.Id,
		Beid:             awards.Beid,
		Ptyid:            awards.Ptyid,
		ActivityId:       awards.ActivityId,
		StartProbability: awards.StartProbability,
		EndProbability:   awards.EndProbability,
		Des:              awards.Des,
		Title:            awards.Title.String,
		Header:           awards.Header,
		Image:            awards.Image.String,
	}
}

func NewGetAwardInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAwardInfoLogic {
	return &GetAwardInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取 问答奖品信息
func (l *GetAwardInfoLogic) GetAwardInfo(in *questions.ActivitiesReq) (*questions.AwardInfoResp, error) {
	AwardInfo, err := l.svcCtx.QuestionsAwardsModel.FindOne(in.Actid)
	if err != nil {
		return nil, err
	}
	return Awards2AwardInfoResp(AwardInfo), nil
}
