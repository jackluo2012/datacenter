package model

import (
	"context"
	"errors"
)

// 内置角色 key
const (
	RoleKeyAdmin  = "admin"  // 超级管理员：全部菜单（含按钮）
	RoleKeyViewer = "viewer" // 只读用户：全部目录+菜单，无按钮
)

// 内置账号
const (
	// SeedAdminUsername 内置管理员登录名
	SeedAdminUsername = "admin"
	// SeedAdminPassword 内置管理员初始密码（bcrypt 后入库，首登后应修改）
	SeedAdminPassword = "admin123"
)

// seedMenu 菜单种子节点定义（递归树）
type seedMenu struct {
	Name      string // 菜单名（同父下唯一，幂等定位键）
	Path      string // 路由路径（目录 /system；子菜单相对段 user；按钮留空）
	Component string // 组件路径（如 system/UserList → web/src/views/system/UserList.vue）
	Icon      string // 图标
	MenuType  int64  // 1.目录 2.菜单 3.按钮
	Perm      string // 按钮权限标识（模块:动作）
	Sort      int64
	Children  []seedMenu
}

// seedAcc 种子收集器：汇总菜单 id，供角色绑定
type seedAcc struct {
	all  []int64 // 全部菜单（admin 角色）
	view []int64 // 目录+菜单（viewer 角色，不含按钮）
}

// SeedRBAC 播种完整 RBAC 种子（幂等可重入：菜单按 父+名 定位存在跳过，
// 角色/账号存在跳过，绑定 insert ignore）：
//  1. 系统管理目录（用户/角色/菜单三页）+ 各页按钮权限
//  2. 数据中台 13 组目录+菜单（数据源/元数据/.../运维）+ 顶层仪表盘 + 核心按钮权限
//  3. 内置角色 admin（全部菜单）/ viewer（目录+菜单，无按钮）
//  4. 内置账号 admin/admin123（bcrypt，绑定 admin 角色）
//
// 并发说明：首查竞态下同名菜单可能重复插入（父+名无唯一约束），概率极低，
// 重复行不影响鉴权语义（按 id 关联），此处以简单性换取无锁实现
func SeedRBAC(ctx context.Context, um UserModel, rm RoleModel, mm MenuModel, beid int64) error {
	// 1. 菜单树
	acc := &seedAcc{}
	if err := seedMenus(ctx, mm, beid, 0, menuSeeds(), acc); err != nil {
		return err
	}

	// 2. 内置角色
	adminId, err := seedRole(ctx, rm, beid, RoleKeyAdmin, "超级管理员", "系统内置：全部菜单与按钮权限")
	if err != nil {
		return err
	}
	viewerId, err := seedRole(ctx, rm, beid, RoleKeyViewer, "只读用户", "系统内置：全部目录与菜单，无按钮权限")
	if err != nil {
		return err
	}

	// 3. 角色菜单绑定（insert ignore 幂等，不覆盖运营期手工调整）
	if err = rm.AddRoleMenus(ctx, adminId, acc.all); err != nil {
		return err
	}
	if err = rm.AddRoleMenus(ctx, viewerId, acc.view); err != nil {
		return err
	}

	// 4. 内置账号 admin/admin123（角色 admin）
	return seedAdminUser(ctx, um, adminId, beid)
}

// seedMenus 递归播种菜单：按 父+名 定位，存在跳过、缺失补插
func seedMenus(ctx context.Context, mm MenuModel, beid, parentId int64, seeds []seedMenu, acc *seedAcc) error {
	for _, s := range seeds {
		var id int64
		exist, err := mm.FindOneByName(ctx, beid, parentId, s.Name)
		switch {
		case err == nil:
			id = exist.Id // 幂等：已存在跳过
		case errors.Is(err, ErrNotFound):
			id, err = mm.Insert(ctx, nil, &Menu{
				ParentId:  parentId,
				Name:      s.Name,
				Path:      s.Path,
				Component: s.Component,
				Sort:      s.Sort,
				Icon:      s.Icon,
				MenuType:  s.MenuType,
				Perm:      s.Perm,
				Visible:   1,
				Beid:      beid,
			})
			if err != nil {
				return err
			}
		default:
			return err
		}
		acc.all = append(acc.all, id)
		if s.MenuType != MenuTypeButton {
			acc.view = append(acc.view, id) // viewer 可见：目录+菜单
		}
		if len(s.Children) > 0 {
			if err = seedMenus(ctx, mm, beid, id, s.Children, acc); err != nil {
				return err
			}
		}
	}
	return nil
}

// seedRole 播种单个内置角色（按 beid+role_key 定位，存在跳过）
func seedRole(ctx context.Context, rm RoleModel, beid int64, key, name, remark string) (int64, error) {
	exist, err := rm.FindByKey(ctx, beid, key)
	switch {
	case err == nil:
		return exist.Id, nil
	case errors.Is(err, ErrNotFound):
		return rm.Insert(ctx, nil, &Role{
			RoleKey:  key,
			RoleName: name,
			Status:   StatusEnabled,
			Remark:   remark,
			Beid:     beid,
		})
	default:
		return 0, err
	}
}

// seedAdminUser 播种内置账号 admin/admin123（bcrypt；存在则仅幂等补绑 admin 角色）
func seedAdminUser(ctx context.Context, um UserModel, adminRoleId, beid int64) error {
	exist, err := um.FindByUsername(ctx, SeedAdminUsername)
	switch {
	case err == nil:
		// 账号已存在：不重置密码，仅补绑角色
		return um.AddUserRoles(ctx, exist.Id, []int64{adminRoleId})
	case errors.Is(err, ErrNotFound):
	default:
		return err
	}
	hash, err := HashPassword(SeedAdminPassword)
	if err != nil {
		return err
	}
	userId, err := um.Insert(ctx, nil, &User{
		Username: SeedAdminUsername,
		Password: hash,
		Nickname: "超级管理员",
		Status:   StatusEnabled,
		Remark:   "系统内置管理员",
		Beid:     beid,
	})
	if err != nil {
		return err
	}
	return um.SetUserRoles(ctx, nil, userId, []int64{adminRoleId})
}

// menuSeeds 完整菜单种子树（对标 RuoYi：目录/菜单/按钮三级，perm 为 模块:动作 标识）
func menuSeeds() []seedMenu {
	seeds := []seedMenu{
		// ===== 系统管理 =====
		{
			Name: "系统管理", Path: "/system", Icon: "Setting", MenuType: MenuTypeDir, Sort: 1,
			Children: []seedMenu{
				{
					Name: "用户管理", Path: "user", Component: "system/UserList", Icon: "User", MenuType: MenuTypeMenu, Sort: 1,
					Children: buttonSeeds("sys:user:",
						buttonSeed("用户查询", "list", 1),
						buttonSeed("用户新增", "create", 2),
						buttonSeed("用户修改", "update", 3),
						buttonSeed("用户删除", "delete", 4),
						buttonSeed("重置密码", "resetPwd", 5),
					),
				},
				{
					Name: "角色管理", Path: "role", Component: "system/RoleList", Icon: "Avatar", MenuType: MenuTypeMenu, Sort: 2,
					Children: buttonSeeds("sys:role:",
						buttonSeed("角色查询", "list", 1),
						buttonSeed("角色新增", "create", 2),
						buttonSeed("角色修改", "update", 3),
						buttonSeed("角色删除", "delete", 4),
					),
				},
				{
					Name: "菜单管理", Path: "menu", Component: "system/MenuList", Icon: "Menu", MenuType: MenuTypeMenu, Sort: 3,
					Children: buttonSeeds("sys:menu:",
						buttonSeed("菜单查询", "list", 1),
						buttonSeed("菜单新增", "create", 2),
						buttonSeed("菜单修改", "update", 3),
						buttonSeed("菜单删除", "delete", 4),
					),
				},
				{
					Name: "操作日志", Path: "oplog", Component: "system/OpLogList", Icon: "Document", MenuType: MenuTypeMenu, Sort: 4,
				},
				{
					Name: "登录日志", Path: "loginlog", Component: "system/LoginLogList", Icon: "Key", MenuType: MenuTypeMenu, Sort: 5,
				},
			},
		},
		// ===== 仪表盘（顶层菜单） =====
		{Name: "仪表盘", Path: "/dashboard", Component: "Dashboard", Icon: "Odometer", MenuType: MenuTypeMenu, Sort: 10},
		// ===== 数据中台 13 组目录+菜单 =====
		dataModule("数据源", "datasource", "DataSourceList", "Coin", 11),
		dataModule("元数据", "metadata", "MetadataList", "Collection", 12,
			buttonSeed("元数据采集", "collect", 5)),
		dataModule("数据服务", "datasvc", "List", "Promotion", 13,
			buttonSeed("服务发布", "publish", 5)),
		dataModule("资产目录", "catalog", "List", "Files", 14,
			buttonSeed("资产上架", "publish", 5)),
		dataModule("数据质量", "quality", "List", "Checked", 15,
			buttonSeed("质量扫描", "run", 5)),
		dataModule("指标管理", "metric", "List", "TrendCharts", 16),
		dataModule("标签管理", "tag", "List", "PriceTag", 17,
			buttonSeed("标签圈群", "circle", 5)),
		dataModule("数据标准", "standard", "List", "Document", 18),
		dataModule("主数据", "mdm", "List", "Box", 19),
		dataModule("调度管理", "job", "List", "Timer", 20,
			buttonSeed("任务执行", "run", 5)),
		dataModule("数据IDE", "ide", "List", "EditPen", 21,
			buttonSeed("脚本运行", "run", 5)),
		dataModule("AI 问数", "ai", "AiChat", "ChatDotRound", 22,
			buttonSeed("AI 提问", "ask", 5)),
		dataModule("数据运维", "ops", "List", "Monitor", 23),
		dataModule("数仓建模", "theme", "ThemeList", "OfficeBuilding", 24,
			buttonSeed("模型部署", "deploy", 5),
			buttonSeed("物化刷新", "refresh", 6)),
		dataModule("数据安全", "security", "SecurityList", "Lock", 25,
			buttonSeed("分级保存", "level", 5),
			buttonSeed("实体打标", "tag", 6),
			buttonSeed("规则保存", "rule", 7),
			buttonSeed("脱敏预览", "preview", 8),
			buttonSeed("授权保存", "grant", 9)),
	}
	return seeds
}

// dataModule 生成数据中台模块种子：目录（/模块）→ 菜单（模块/组件）→ 核心按钮
// （模块:list/create/update/delete）+ 模块专属扩展按钮（sort 从 5 起）
func dataModule(name, mod, comp, icon string, sort int64, extras ...seedMenu) seedMenu {
	buttons := []seedMenu{
		buttonSeed(name+"查询", "list", 1),
		buttonSeed(name+"新增", "create", 2),
		buttonSeed(name+"修改", "update", 3),
		buttonSeed(name+"删除", "delete", 4),
	}
	buttons = append(buttons, extras...)
	return seedMenu{
		Name: name, Path: "/" + mod, Icon: icon, MenuType: MenuTypeDir, Sort: sort,
		Children: []seedMenu{
			{
				Name: name + "管理", Path: mod, Component: mod + "/" + comp, Icon: icon, MenuType: MenuTypeMenu, Sort: 1,
				Children: buttonSeeds(mod+":", buttons...),
			},
		},
	}
}

// buttonSeeds 为按钮补全权限标识前缀（模块:动作）
func buttonSeeds(prefix string, buttons ...seedMenu) []seedMenu {
	for i := range buttons {
		buttons[i].Perm = prefix + buttons[i].Perm
		buttons[i].MenuType = MenuTypeButton
	}
	return buttons
}

// buttonSeed 按钮种子节点（Perm 存动作，由 buttonSeeds 拼模块前缀）
func buttonSeed(name, action string, sort int64) seedMenu {
	return seedMenu{
		Name:     name,
		Perm:     action,
		MenuType: MenuTypeButton,
		Sort:     sort,
	}
}
