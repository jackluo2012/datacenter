package logic

import (
	"context"

	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUsersLogic) ListUsers(in *sys.SysUserListReq) (*sys.SysUserListReply, error) {
	// todo: add your logic here and delete this line

	return &sys.SysUserListReply{}, nil
}
