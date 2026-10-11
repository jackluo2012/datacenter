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

type SysGetinfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysGetinfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysGetinfoLogic {
	return &SysGetinfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 当前登录用户信息(uid/角色/按钮权限/菜单树): 网关从 JWT ctx 取 uid/beid 传入,
// 替代样板前端 auth store 会话, 转发 sys.rpc GetInfo
func (l *SysGetinfoLogic) SysGetinfo() (*types.SysGetInfoResp, error) {
	reply, err := l.svcCtx.SysRpc.GetInfo(l.ctx, &sysclient.GetInfoReq{
		Uid:  ctxUid(l.ctx),
		Beid: ctxBeid(l.ctx),
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	menus := make([]types.SysMenuInfo, 0, len(reply.GetMenus()))
	for _, item := range reply.GetMenus() {
		menus = append(menus, *sysMenuToTypes(item))
	}
	roles := make([]string, 0, len(reply.GetRoles()))
	roles = append(roles, reply.GetRoles()...)
	perms := make([]string, 0, len(reply.GetPerms()))
	perms = append(perms, reply.GetPerms()...)
	return &types.SysGetInfoResp{
		Uid:      reply.GetUid(),
		Username: reply.GetUsername(),
		Nickname: reply.GetNickname(),
		Roles:    roles,
		Perms:    perms,
		Menus:    menus,
	}, nil
}
