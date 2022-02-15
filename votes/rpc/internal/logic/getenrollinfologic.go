package logic

import (
	"context"
	"encoding/json"

	"datacenter/votes/rpc/internal/svc"
	"datacenter/votes/rpc/votes"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEnrollInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEnrollInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEnrollInfoLogic {
	return &GetEnrollInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取作品信息
func (l *GetEnrollInfoLogic) GetEnrollInfo(in *votes.EnrollInfoReq) (*votes.EnrollResp, error) {
	vote, err := l.svcCtx.AppEnrollModel.FindOne(in.Aeid)
	if err != nil {
		logx.Info("EnrollNotExist==", EnrollNotExist)
		return nil, EnrollNotExist
	}
	image := make([]string, 0)
	err = json.Unmarshal([]byte(vote.Images), &image)
	if err != nil {
		return nil, err
	}
	//增加 浏览量和
	l.svcCtx.AppEnrollModel.IncrView(in.Aeid)
	l.svcCtx.AppVotesActivityModel.IncrView(in.Actid)
	return &votes.EnrollResp{
		Ptyid:     vote.Ptyid,
		Uid:       vote.Uid,
		Beid:      vote.Beid,
		Auid:      vote.Auid,
		Actid:     vote.Actid,
		Name:      vote.Name,
		Images:    image,
		Descr:     vote.Descr,
		Aeid:      vote.Aeid,
		Address:   vote.Address,
		Viewcount: vote.Viewcount,
		Votecount: vote.Votecount,
	}, nil
}
