package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"
	"datacenter/votes/rpc/votesclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type EnrollLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEnrollLogic(ctx context.Context, svcCtx *svc.ServiceContext) EnrollLogic {
	return EnrollLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EnrollLogic) Enroll(req types.EnrollReq) (*types.EnrollResp, error) {
	//获取用户的值

	images := make([]string, 0)
	for _, val := range req.Images {
		images = append(images, shared.CDN_SSO_Qiuniu+val)
	}

	jsonStu, err := json.Marshal(images)
	if err != nil {
		return nil, err
	}
	auid, _ := json.Number(fmt.Sprintf("%v", l.ctx.Value("auid"))).Int64()
	uid, _ := json.Number(fmt.Sprintf("%v", l.ctx.Value("uid"))).Int64()
	ptyid, _ := json.Number(fmt.Sprintf("%v", l.ctx.Value("ptyid"))).Int64()
	beid, _ := json.Number(fmt.Sprintf("%v", l.ctx.Value("beid"))).Int64()
	if auid == 0 && uid == 0 {
		return nil, shared.ErrorUserNotFound
	}
	logx.Info("ptyid=", ptyid, "beid=", beid)

	relp, err := l.svcCtx.VotesRpc.Enroll(l.ctx, &votesclient.EnrollReq{
		Auid:    auid,
		Uid:     uid,
		Ptyid:   ptyid,
		Beid:    beid,
		Name:    req.Name,
		Address: req.Address,
		Descr:   req.Descr,
		Images:  string(jsonStu),
		Actid:   req.Actid.Actid,
	})
	if err != nil {
		return nil, err
	}
	return &types.EnrollResp{
		Aeid: relp.Aeid,
	}, nil
}
