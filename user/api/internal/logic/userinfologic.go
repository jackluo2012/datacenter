package logic

import (
	"context"
	"strconv"

	"datacenter/user/api/internal/svc"
	"datacenter/user/api/internal/types"
	"datacenter/user/model"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *UserInfoLogic) UserInfo(userId string) (*types.UserReply, error) {

	userInt, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return nil, err
	}

	userInfo, err := l.svcCtx.UserModel.FindOne(userInt)
	switch err {
	case nil:
		return &types.UserReply{
			Uid:      userInfo.Id,
			Username: userInfo.Username,
			Mobile:   userInfo.Mobile,
		}, nil
	case model.ErrNotFound:
		return nil, errorUserNotFound
	default:
		return nil, err
	}
}
