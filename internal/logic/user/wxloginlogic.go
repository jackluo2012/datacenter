package logic

import (
	"context"

	"datacenter/internal/logic"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/user/rpc/userclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type WxloginLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	datacenter logic.DatacenterLogic
}

func NewWxloginLogic(ctx context.Context, svcCtx *svc.ServiceContext) WxloginLogic {
	return WxloginLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		datacenter: logic.NewDatacenterLogic(ctx, svcCtx),
	}
}

func (l *WxloginLogic) Wxlogin(req types.WxLoginReq) (*types.LoginAppUser, error) {

	reply, err := l.svcCtx.UserRpc.SnsLogin(l.ctx, &userclient.AppConfigReq{Beid: req.Beid, Ptyid: req.Ptyid, Code: req.Code})
	if err != nil {
		return nil, err
	}
	token, err := l.datacenter.GetJwtToken(types.AppUser{
		Auid:     reply.Auid,
		Beid:     reply.Beid,
		Ptyid:    reply.Ptyid,
		Nickname: reply.Nickname,
		Openid:   reply.Openid,
	})
	if err != nil {
		return nil, err
	}
	return &types.LoginAppUser{
		Auid:     reply.Auid,
		Beid:     reply.Beid,
		Ptyid:    reply.Ptyid,
		Nickname: reply.Nickname,
		Openid:   reply.Openid,
		Avator:   reply.Avator,
		JwtToken: token,
	}, nil
}
