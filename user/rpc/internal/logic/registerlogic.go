package logic

import (
	"context"

	"datacenter/user/rpc/internal/svc"
	"datacenter/user/rpc/user"

	"datacenter/user/model"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.UserReply, error) {

	_, err := l.svcCtx.UserModel.FindOneByMobile(in.Mobile)
	if err == nil {
		return nil, errorDuplicateMobile
	}

	// 处理加盐算法
	res, err := l.svcCtx.UserModel.Insert(model.BaseMember{
		Mobile:    in.Mobile,
		Password:  in.Password,
		DeletedAt: time.Unix(0, 0),
	})
	if err != nil {
		return nil, err
	}
	uid, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &user.UserReply{
		Uid:    uid,
		Mobile: in.Mobile,
	}, nil
}
