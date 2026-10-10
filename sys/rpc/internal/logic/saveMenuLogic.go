package logic

import (
	"context"

	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveMenuLogic {
	return &SaveMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 菜单
func (l *SaveMenuLogic) SaveMenu(in *sys.SaveMenuReq) (*sys.MenuReply, error) {
	// todo: add your logic here and delete this line

	return &sys.MenuReply{}, nil
}
