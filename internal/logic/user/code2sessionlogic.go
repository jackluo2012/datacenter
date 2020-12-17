package logic

import (
	"context"

	"datacenter/internal/logic"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/user/rpc/userclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type Code2SessionLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	datacenter logic.DatacenterLogic
}

func NewCode2SessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) Code2SessionLogic {
	return Code2SessionLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		datacenter: logic.NewDatacenterLogic(ctx, svcCtx),
	}
}

func (l *Code2SessionLogic) Code2Session(beid, ptyid int64, code string) (lp *types.LoginAppUser, err error) {

	reply, err := l.svcCtx.UserRpc.SnsLogin(l.ctx, &userclient.AppConfigReq{Beid: beid, Ptyid: ptyid, Code: code})
	if err != nil {
		return
	}
	token, err := l.datacenter.GetJwtToken(types.AppUser{
		Auid:     reply.Auid,
		Beid:     reply.Beid,
		Ptyid:    reply.Ptyid,
		Nickname: reply.Nickname,
		Openid:   reply.Openid,
	})
	if err != nil {
		return
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
