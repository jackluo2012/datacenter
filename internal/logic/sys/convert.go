// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package sys

import (
	"context"
	"encoding/json"
	"fmt"

	"datacenter/internal/types"
	sysclient "datacenter/sys/rpc/sysClient"
)

// ---------- proto -> 网关 types 转换 ----------

// sysUserToTypes SysUser proto 转 HTTP 响应结构
func sysUserToTypes(in *sysclient.SysUser) *types.SysUserInfo {
	if in == nil {
		return &types.SysUserInfo{}
	}
	roleIds := make([]int64, 0, len(in.RoleIds))
	roleIds = append(roleIds, in.RoleIds...)
	return &types.SysUserInfo{
		Id:        in.Id,
		Username:  in.Username,
		Nickname:  in.Nickname,
		Mobile:    in.Mobile,
		Email:     in.Email,
		Status:    in.Status,
		Remark:    in.Remark,
		CreatedAt: in.CreatedAt,
		RoleIds:   roleIds,
	}
}

// sysRoleToTypes SysRole proto 转 HTTP 响应结构
func sysRoleToTypes(in *sysclient.SysRole) *types.SysRoleInfo {
	if in == nil {
		return &types.SysRoleInfo{}
	}
	menuIds := make([]int64, 0, len(in.MenuIds))
	menuIds = append(menuIds, in.MenuIds...)
	return &types.SysRoleInfo{
		Id:       in.Id,
		RoleKey:  in.RoleKey,
		RoleName: in.RoleName,
		Status:   in.Status,
		Remark:   in.Remark,
		Beid:     in.Beid,
		MenuIds:  menuIds,
	}
}

// sysMenuToTypes SysMenu proto 转 HTTP 响应结构（递归转换子菜单树）
func sysMenuToTypes(in *sysclient.SysMenu) *types.SysMenuInfo {
	if in == nil {
		return &types.SysMenuInfo{}
	}
	children := make([]types.SysMenuInfo, 0, len(in.Children))
	for _, child := range in.Children {
		children = append(children, *sysMenuToTypes(child))
	}
	return &types.SysMenuInfo{
		Id:        in.Id,
		ParentId:  in.ParentId,
		Name:      in.Name,
		Path:      in.Path,
		Component: in.Component,
		Sort:      in.Sort,
		Icon:      in.Icon,
		MenuType:  in.MenuType,
		Perm:      in.Perm,
		Visible:   in.Visible,
		Beid:      in.Beid,
		Children:  children,
	}
}

// routerBranchToTypes RouterBranch proto 转 HTTP 响应结构（递归转换子路由）
func routerBranchToTypes(in *sysclient.RouterBranch) *types.SysRouterBranch {
	if in == nil {
		return &types.SysRouterBranch{}
	}
	children := make([]types.SysRouterBranch, 0, len(in.Children))
	for _, child := range in.Children {
		children = append(children, *routerBranchToTypes(child))
	}
	return &types.SysRouterBranch{
		Name:      in.Name,
		Path:      in.Path,
		Component: in.Component,
		Icon:      in.Icon,
		Children:  children,
	}
}

// ---------- JWT ctx 取值（网关编排：uid/beid 从 JWT ctx 注入，替代样板的前端 auth store 会话） ----------

// ctxInt64 从 JWT claims ctx 中取整数值（go-zero 注入的 claim 可能是 json.Number/float64）
func ctxInt64(ctx context.Context, key string) int64 {
	val, _ := json.Number(fmt.Sprintf("%v", ctx.Value(key))).Int64()
	return val
}

// ctxUid 当前登录用户 uid（未登录/无 claims 时为 0）
func ctxUid(ctx context.Context) int64 {
	return ctxInt64(ctx, "uid")
}

// ctxBeid 当前登录用户所属应用 beid（无 claims 时为 0）
func ctxBeid(ctx context.Context) int64 {
	return ctxInt64(ctx, "beid")
}

// beidOrCtx 请求未传 beid 时回退取 JWT ctx 中的 beid
func beidOrCtx(reqBeid int64, ctx context.Context) int64 {
	if reqBeid != 0 {
		return reqBeid
	}
	return ctxBeid(ctx)
}
