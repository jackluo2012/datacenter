package logic

import (
	"context"

	"datacenter/internal/logic"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/user/rpc/userclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	datacenterLogic logic.DatacenterLogic
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) LoginLogic {
	return LoginLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		datacenterLogic: logic.NewDatacenterLogic(ctx, svcCtx),
	}
}

func (l *LoginLogic) Login(req types.LoginReq, beid, ptyid int64) (*types.UserReply, error) {

	reply, err := l.svcCtx.UserRpc.Login(l.ctx, &userclient.LoginReq{
		Mobile:   req.Mobile,
		Type:     req.Type,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	jwttoken, err := l.datacenterLogic.GetJwtToken(types.AppUser{Uid: reply.Uid, Ptyid: ptyid, Beid: beid})
	if err != nil {
		return nil, err
	}
	return &types.UserReply{
		Auid:     reply.Auid,
		Beid:     beid,
		Ptyid:    ptyid,
		Uid:      reply.Uid,
		Username: reply.Username,
		Mobile:   reply.Mobile,
		Avator:   reply.Avator,
		JwtToken: jwttoken,
	}, nil
}
