package logic

import (
	"context"
	"encoding/json"

	"datacenter/votes/rpc/internal/svc"
	"datacenter/votes/rpc/votes"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEnrollListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEnrollListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEnrollListLogic {
	return &GetEnrollListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取 作品列表
func (l *GetEnrollListLogic) GetEnrollList(in *votes.ActidReq) (*votes.EnrollListResp, error) {
	//这里可以控制走哪里
	list, err := l.svcCtx.AppEnrollModel.Find(in)
	if err != nil {
		return nil, err
	}

	lists := make([]*votes.EnrollResp, 0)
	for _, vote := range list {
		images := make([]string, 0)
		err := json.Unmarshal([]byte(vote.Images), &images)
		if err != nil {
			return nil, err
		}
		enrollResp := votes.EnrollResp{
			Ptyid:     vote.Ptyid,
			Uid:       vote.Uid,
			Beid:      vote.Beid,
			Auid:      vote.Auid,
			Aeid:      vote.Aeid,
			Actid:     vote.Actid,
			Name:      vote.Name,
			Descr:     vote.Descr,
			Images:    images,
			Address:   vote.Address,
			Viewcount: vote.Viewcount,
			Votecount: vote.Votecount,
		}
		logx.Info("enrollResp.Aeid===", enrollResp.Aeid)
		lists = append(lists, &enrollResp)
	}

	return &votes.EnrollListResp{
		Data: lists,
	}, nil
}
