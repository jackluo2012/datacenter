package logic

import (
	"context"
	"time"

	"datacenter/votes/model"
	"datacenter/votes/rpc/internal/svc"
	"datacenter/votes/rpc/votes"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnrollLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnrollLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnrollLogic {
	return &EnrollLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EnrollLogic) Enroll(in *votes.EnrollReq) (*votes.EnrollResp, error) {

	//检查 用户是否重复报名
	count, err := l.svcCtx.AppEnrollModel.GetActIdOrAcidExist(in.Actid, in.Uid, in.Auid)
	if err != nil {
		return nil, YouHaveSignedUp
	}
	if count > 0 {
		//return nil, YouHaveSignedUp
	}
	// 获取 投票活动的信息
	activit, err := l.svcCtx.AppVotesActivityModel.FindOne(in.Actid)
	if err != nil {
		return nil, ActivityDoesNotExist
	}
	// 检查 活动的状态
	if activit.Status != 1 {
		return nil, ActivityDoesNotOpen
	}
	//检查 活动是否开始报名
	nowtime := time.Now().Unix()
	if activit.StartDate > nowtime {
		return nil, ActivityDoesNotStart
	}
	//时间是否过了
	if activit.EnrollDate < nowtime {
		//return nil, RegistrationActivityDoesNotEnrollEND
	}
	//活动结束
	if activit.EndDate < nowtime {
		return nil, ActivityEnd
	}
	resl, err := l.svcCtx.AppEnrollModel.Insert(model.AppEnroll{
		Ptyid:   in.Ptyid,
		Beid:    in.Beid,
		Actid:   in.Actid,
		Uid:     in.Uid,
		Auid:    in.Auid,
		Name:    in.Name,
		Status:  1,
		Address: in.Address,
		Descr:   in.Descr,
		Images:  in.Images,
	})
	if err != nil {
		return nil, EnrollFalt
	}
	id, err := resl.LastInsertId()
	if err != nil {
		return nil, EnrollFalt
	}
	// 报名成功了 活动报名 +1
	l.svcCtx.AppVotesActivityModel.IncrEnroll(in.Actid)

	return &votes.EnrollResp{
		Aeid:  id,
		Ptyid: in.Ptyid,
		Beid:  in.Beid,
		Actid: in.Actid,
		Auid:  in.Auid,
	}, nil
}
