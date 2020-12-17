package logic

import (
	"context"

	"datacenter/internal/logic"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/user/rpc/userclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	datacenterLogic logic.DatacenterLogic
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) RegisterLogic {
	return RegisterLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		datacenterLogic: logic.NewDatacenterLogic(ctx, svcCtx),
	}
}

func (l *RegisterLogic) Register(req types.RegisterReq) (*types.UserReply, error) {
	reply, err := l.svcCtx.UserRpc.Register(l.ctx, &userclient.RegisterReq{
		Mobile:   req.Mobile,
		Password: req.Password,
		Smscode:  req.Smscode,
	})
	if err != nil {
		return nil, err
	}
	jwttoken, err := l.datacenterLogic.GetJwtToken(types.AppUser{Uid: reply.Uid})
	if err != nil {
		return nil, err
	}
	return &types.UserReply{
		Auid:     reply.Auid,
		Beid:     reply.Beid,
		Ptyid:    reply.Ptyid,
		Uid:      reply.Uid,
		Username: reply.Username,
		Mobile:   reply.Mobile,
		Avator:   reply.Avator,
		JwtToken: jwttoken,
	}, nil
}
