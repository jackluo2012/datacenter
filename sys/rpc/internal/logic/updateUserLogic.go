package logic

import (
	"context"

	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserLogic) UpdateUser(in *sys.UpdateSysUserReq) (*sys.SysUserReply, error) {
	// todo: add your logic here and delete this line

	return &sys.SysUserReply{}, nil
}
