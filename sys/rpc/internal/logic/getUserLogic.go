package logic

import (
	"context"
	"errors"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户详情：用户信息 + role_ids（不含密码哈希）
func (l *GetUserLogic) GetUser(in *sys.IdReq) (*sys.SysUserReply, error) {
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	roleIds, err := l.svcCtx.UserModel.GetRoleIds(l.ctx, user.Id)
	if err != nil {
		return nil, err
	}
	return &sys.SysUserReply{User: userToPb(user, roleIds)}, nil
}
