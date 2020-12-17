package logic

import (
	"context"
	"time"

	"datacenter/user/api/internal/svc"
	"datacenter/user/api/internal/types"

	"datacenter/user/model"

	"github.com/tal-tech/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) RegisterLogic {
	return RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req types.RegisterReq) error {

	_, err := l.svcCtx.UserModel.FindOneByMobile(req.Mobile)
	if err == nil {
		return errorDuplicateMobile
	}

	// 处理加盐算法

	_, err = l.svcCtx.UserModel.Insert(model.BaseMember{
		Username:  req.Username,
		Mobile:    req.Mobile,
		Password:  req.Password,
		DeletedAt: time.Unix(0, 0),
	})
	return err

}
