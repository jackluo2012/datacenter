package logic

import (
	"context"

	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type MenuTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuTreeLogic {
	return &MenuTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MenuTreeLogic) MenuTree(in *sys.IdReq) (*sys.MenuTreeReply, error) {
	// todo: add your logic here and delete this line

	return &sys.MenuTreeReply{}, nil
}
