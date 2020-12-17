package logic

import (
	"context"

	"datacenter/votes/rpc/internal/svc"
	"datacenter/votes/rpc/votes"

	"github.com/tal-tech/go-zero/core/logx"
)

type GetActivityInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetActivityInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityInfoLogic {
	return &GetActivityInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取 活动信息
func (l *GetActivityInfoLogic) GetActivityInfo(in *votes.ActInfoReq) (*votes.ActInfoResp, error) {
	activity, err := l.svcCtx.AppVotesActivityModel.FindOne(in.Actid)
	if err != nil {
		return nil, err
	}
	list := &votes.ActInfoResp{
		Actid:       activity.Actid,
		Beid:        activity.Beid,
		Descr:       activity.Descr,
		EndDate:     activity.EndDate,
		StartDate:   activity.StartDate,
		EnrollDate:  activity.EnrollDate,
		Type:        activity.Type,
		Title:       activity.Title,
		Viewcount:   activity.Viewcount,
		Votecount:   activity.Votecount,
		Enrollcount: activity.Enrollcount,
		Num:         activity.Num,
	}
	return list, nil
}
