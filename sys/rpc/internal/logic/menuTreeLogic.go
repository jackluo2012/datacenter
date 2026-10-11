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

// 全量菜单树（管理用）：首次为空自动播种 RBAC 种子（幂等），全量内存建树
func (l *MenuTreeLogic) MenuTree(in *sys.IdReq) (*sys.MenuTreeReply, error) {
	list, err := l.svcCtx.MenuModel.FindAll(l.ctx, in.Beid)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		// 种子初始化：仅菜单为空时执行，父+名定位存在跳过，可重入
		if err = seedRBAC(l.ctx, l.svcCtx, in.Beid); err != nil {
			return nil, err
		}
		if list, err = l.svcCtx.MenuModel.FindAll(l.ctx, in.Beid); err != nil {
			return nil, err
		}
	}
	tree := BuildMenuTree(menusToPb(list))
	return &sys.MenuTreeReply{List: tree}, nil
}
