package logic

import (
	"datacenter/sys/model"
	"datacenter/sys/rpc/sys"
)

// userToPb 用户行转 pb（不含密码哈希，防外透；roleIds 为用户角色 id 集）
func userToPb(u *model.User, roleIds []int64) *sys.SysUser {
	if u == nil {
		return nil
	}
	return &sys.SysUser{
		Id:        u.Id,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Mobile:    u.Mobile,
		Email:     u.Email,
		Status:    int32(u.Status),
		Remark:    u.Remark,
		CreatedAt: u.CreatedAt.UnixMilli(),
		RoleIds:   roleIds,
	}
}

// roleToPb 角色行转 pb（menuIds 为角色菜单 id 集）
func roleToPb(r *model.Role, menuIds []int64) *sys.SysRole {
	if r == nil {
		return nil
	}
	return &sys.SysRole{
		Id:       r.Id,
		RoleKey:  r.RoleKey,
		RoleName: r.RoleName,
		Status:   int32(r.Status),
		Remark:   r.Remark,
		Beid:     r.Beid,
		MenuIds:  menuIds,
	}
}

// menuToPb 菜单行转 pb
func menuToPb(m *model.Menu) *sys.SysMenu {
	if m == nil {
		return nil
	}
	return &sys.SysMenu{
		Id:        m.Id,
		ParentId:  m.ParentId,
		Name:      m.Name,
		Path:      m.Path,
		Component: m.Component,
		Sort:      int32(m.Sort),
		Icon:      m.Icon,
		MenuType:  int32(m.MenuType),
		Perm:      m.Perm,
		Visible:   int32(m.Visible),
		Beid:      m.Beid,
	}
}

// menusToPb 菜单行列表转 pb 列表
func menusToPb(list []*model.Menu) []*sys.SysMenu {
	out := make([]*sys.SysMenu, 0, len(list))
	for _, m := range list {
		out = append(out, menuToPb(m))
	}
	return out
}

// dedupeInt64s 角色/菜单 id 集去重（保序，过滤非正数；透传给 model 层整存语义前统一清洗）
func dedupeInt64s(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	uniq := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	return uniq
}
