package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/user/rpc/userclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserInfoLogic {
	return UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoLogic) UserInfo(req types.UserReq) (*types.UserReply, error) {
	reply, err := l.svcCtx.UserRpc.UserInfo(l.ctx, &userclient.UserReq{})
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
	}, nil
}
