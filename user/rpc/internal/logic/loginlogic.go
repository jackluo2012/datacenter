package logic

import (
	"context"

	"datacenter/user/model"
	"datacenter/user/rpc/internal/svc"
	"datacenter/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginReq) (*user.UserReply, error) {

	// 忽略逻辑校验
	userInfo, err := l.svcCtx.UserModel.FindOneByMobile(in.Mobile)
	switch err {
	case nil:
		if userInfo.Password != in.Password {
			return nil, errorIncorrectPassword
		}
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
