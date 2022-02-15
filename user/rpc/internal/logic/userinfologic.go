package logic

import (
	"context"

	"datacenter/user/model"
	"datacenter/user/rpc/internal/svc"
	"datacenter/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserInfoLogic) UserInfo(in *user.UserReq) (*user.UserReply, error) {

	// 忽略逻辑校验
	userInfo, err := l.svcCtx.UserModel.FindOne(in.Uid)
	switch err {
	case nil:
		return &user.UserReply{
			Uid:      userInfo.Id,
			Username: userInfo.Username,
			Mobile:   userInfo.Mobile,
		}, nil
	case model.ErrNotFound:
		return nil, errorUsernameUnRegister
	default:
		return nil, err
	}
}
