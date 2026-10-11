package logic

import "errors"

// 业务错误（对齐 standard/security 服务写法：errors.New 中文语义错误，调用方用 errors.Is 判断）
var (
	// ErrIdRequired 通用 id 缺失
	ErrIdRequired = errors.New("id 不能为空")
	// ErrNameRequired 名称缺失（角色名/菜单名）
	ErrNameRequired = errors.New("name 不能为空")
	// ErrUsernameRequired 登录名缺失
	ErrUsernameRequired = errors.New("username 不能为空")
	// ErrPasswordRequired 密码缺失
	ErrPasswordRequired = errors.New("password 不能为空")
	// ErrUsernameDuplicate 登录名已存在
	ErrUsernameDuplicate = errors.New("用户名已存在")
	// ErrUserNotFound 用户不存在（含跨租户不可见）
	ErrUserNotFound = errors.New("用户不存在")
	// ErrRoleNotFound 角色不存在（含跨租户不可见）
	ErrRoleNotFound = errors.New("角色不存在")
	// ErrRoleKeyRequired 角色权限字符缺失
	ErrRoleKeyRequired = errors.New("role_key 不能为空")
	// ErrRoleKeyDuplicate 角色权限字符已存在（同 beid 下）
	ErrRoleKeyDuplicate = errors.New("角色权限字符 role_key 已存在")
	// ErrRoleInUse 角色已分配用户，禁止删除
	ErrRoleInUse = errors.New("角色已分配用户，禁止删除")
	// ErrMenuNotFound 菜单不存在（含跨租户不可见）
	ErrMenuNotFound = errors.New("菜单不存在")
	// ErrMenuTypeInvalid menu_type 非法（仅支持 1.目录 2.菜单 3.按钮）
	ErrMenuTypeInvalid = errors.New("menu_type 非法：仅支持 1.目录 2.菜单 3.按钮")
	// ErrMenuPermInvalid perm 格式非法：需为 模块:动作（如 sys:user:create），小写字母数字下划线分段
	ErrMenuPermInvalid = errors.New("perm 非法：需为 模块:动作 格式（如 sys:user:create）")
	// ErrMenuParentInvalid 父节点非法（不存在/类型不符/形成环）
	ErrMenuParentInvalid = errors.New("父节点非法：不存在、类型不符或形成环")
	// ErrMenuHasChildren 存在子菜单，禁止删除
	ErrMenuHasChildren = errors.New("存在子菜单，禁止删除")
	// ErrMenuInUse 菜单已绑定角色，禁止删除
	ErrMenuInUse = errors.New("菜单已绑定角色，禁止删除")
)
