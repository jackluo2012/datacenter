// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package sys

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	sysclient "datacenter/sys/rpc/sysClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysMenuSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuSaveLogic {
	return &SysMenuSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存菜单(目录/菜单/按钮): Id=0 新建否则更新, beid 缺省取 JWT ctx, 转发 sys.rpc SaveMenu
func (l *SysMenuSaveLogic) SysMenuSave(req *types.SysMenuSaveReq) (*types.SysMenuResp, error) {
	reply, err := l.svcCtx.SysRpc.SaveMenu(l.ctx, &sysclient.SaveMenuReq{
		Id:        req.Id,
		ParentId:  req.ParentId,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Sort:      req.Sort,
		Icon:      req.Icon,
		MenuType:  req.MenuType,
		Perm:      req.Perm,
		Visible:   req.Visible,
		Beid:      beidOrCtx(req.Beid, l.ctx),
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.SysMenuResp{Menu: *sysMenuToTypes(reply.GetMenu())}, nil
}
