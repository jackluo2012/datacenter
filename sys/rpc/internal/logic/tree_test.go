package logic

import (
	"reflect"
	"testing"

	"datacenter/sys/rpc/sys"
)

// menu 测试辅助：构造 pb 菜单节点
func menu(id, parentId int64, name string, menuType int32, perm string) *sys.SysMenu {
	return &sys.SysMenu{Id: id, ParentId: parentId, Name: name, MenuType: menuType, Perm: perm}
}

// idsOf 提取节点 id（顺序敏感）
func idsOf(menus []*sys.SysMenu) []int64 {
	ids := make([]int64, 0, len(menus))
	for _, m := range menus {
		ids = append(ids, m.GetId())
	}
	return ids
}

// findChild 按名查找直接子节点
func findChild(m *sys.SysMenu, name string) *sys.SysMenu {
	for _, c := range m.GetChildren() {
		if c.GetName() == name {
			return c
		}
	}
	return nil
}

// TestBuildMenuTreeMultiLevel 多级建树：目录→菜单→按钮三级挂载，层级与顺序正确
func TestBuildMenuTreeMultiLevel(t *testing.T) {
	menus := []*sys.SysMenu{
		menu(1, 0, "系统管理", menuTypeDir, ""),
		menu(2, 1, "用户管理", menuTypeMenu, ""),
		menu(3, 2, "用户新增", menuTypeButton, "sys:user:create"),
		menu(4, 2, "用户删除", menuTypeButton, "sys:user:delete"),
		menu(5, 1, "角色管理", menuTypeMenu, ""),
		menu(6, 0, "仪表盘", menuTypeMenu, ""),
	}
	tree := BuildMenuTree(menus)
	if got := idsOf(tree); !reflect.DeepEqual(got, []int64{1, 6}) {
		t.Fatalf("根应为 [1 系统管理 6 仪表盘]，实际 %v", got)
	}
	sysDir := tree[0]
	if got := idsOf(sysDir.GetChildren()); !reflect.DeepEqual(got, []int64{2, 5}) {
		t.Fatalf("系统管理子级应为 [2 用户管理 5 角色管理]，实际 %v", got)
	}
	userMenu := findChild(sysDir, "用户管理")
	if userMenu == nil {
		t.Fatalf("用户管理应挂在系统管理下")
	}
	if got := idsOf(userMenu.GetChildren()); !reflect.DeepEqual(got, []int64{3, 4}) {
		t.Fatalf("用户管理按钮应为 [3 4]，实际 %v", got)
	}
}

// TestBuildMenuTreeOrphan 孤儿挂根：父节点缺失的节点兜底挂根，不丢失
func TestBuildMenuTreeOrphan(t *testing.T) {
	menus := []*sys.SysMenu{
		menu(1, 0, "根", menuTypeDir, ""),
		menu(2, 999, "孤儿", menuTypeMenu, ""), //父 999 不存在
		menu(3, 2, "孤儿子页", menuTypeMenu, ""), //父存在但自身链可达
	}
	tree := BuildMenuTree(menus)
	if got := idsOf(tree); !reflect.DeepEqual(got, []int64{1, 2}) {
		t.Fatalf("孤儿应挂根：根为 [1 2]，实际 %v", got)
	}
	orphan := tree[1]
	if got := idsOf(orphan.GetChildren()); !reflect.DeepEqual(got, []int64{3}) {
		t.Fatalf("孤儿子节点应正常挂父，实际 %v", got)
	}
}

// TestBuildMenuTreeCycle 环防护：互指成环的节点全部兜底挂根，节点不丢失、遍历不死循环
func TestBuildMenuTreeCycle(t *testing.T) {
	menus := []*sys.SysMenu{
		menu(1, 2, "A", menuTypeDir, ""), //A→B
		menu(2, 1, "B", menuTypeDir, ""), //B→A（成环）
		menu(3, 2, "C", menuTypeMenu, ""), //C 挂在环内节点上
		menu(4, 0, "正常", menuTypeDir, ""),
		menu(5, 5, "自环", menuTypeDir, ""), //自环
	}
	tree := BuildMenuTree(menus)
	//环内 A/B（互指双方都在环上）、环上 C、自环 5、正常节点 4 全部挂根
	if got := idsOf(tree); !reflect.DeepEqual(got, []int64{1, 2, 3, 4, 5}) {
		t.Fatalf("环/自环/环上节点应全部挂根：实际 %v", got)
	}
	//环内节点互相不再作为父子挂载（B 不在 A 的 children 里）
	if got := idsOf(tree[0].GetChildren()); len(got) != 0 {
		t.Fatalf("环内节点 children 应为空（互不挂载），实际 %v", got)
	}
	//遍历整树不panic、节点总数不丢
	count := 0
	var walk func(menus []*sys.SysMenu)
	walk = func(ms []*sys.SysMenu) {
		for _, m := range ms {
			count++
			if count > 100 {
				t.Fatalf("建树出现死循环")
			}
			walk(m.GetChildren())
		}
	}
	walk(tree)
	if count != 5 {
		t.Fatalf("5 个节点应全部可达，实际 %d", count)
	}
}

// TestCollectPerms perm 集合计算：仅收集按钮非空标识，去重保序，跨层收集
func TestCollectPerms(t *testing.T) {
	tree := []*sys.SysMenu{
		{Id: 1, Name: "系统管理", MenuType: menuTypeDir, Children: []*sys.SysMenu{
			{Id: 2, Name: "用户管理", MenuType: menuTypeMenu, Children: []*sys.SysMenu{
				{Id: 3, Name: "用户查询", MenuType: menuTypeButton, Perm: "sys:user:list"},
				{Id: 4, Name: "用户新增", MenuType: menuTypeButton, Perm: "sys:user:create"},
				{Id: 5, Name: "空权限按钮", MenuType: menuTypeButton, Perm: ""}, //空标识忽略
			}},
			{Id: 6, Name: "角色管理", MenuType: menuTypeMenu, Children: []*sys.SysMenu{
				{Id: 7, Name: "重复查询", MenuType: menuTypeButton, Perm: "sys:user:list"}, //重复去重
			}},
		}},
		{Id: 8, Name: "带权限的菜单", MenuType: menuTypeMenu, Perm: "dashboard:view"}, //非按钮类型也收集（perm 为权威来源）
	}
	perms := CollectPerms(tree)
	want := []string{"sys:user:list", "sys:user:create", "dashboard:view"}
	if !reflect.DeepEqual(perms, want) {
		t.Fatalf("perms = %v, want %v", perms, want)
	}
	if got := CollectPerms(nil); len(got) != 0 {
		t.Fatalf("空树 perms 应为空，实际 %v", got)
	}
}

// TestStripButtons 按钮剔除：目录+菜单保留，按钮整支剔除，其余结构不变
func TestStripButtons(t *testing.T) {
	tree := []*sys.SysMenu{
		{Id: 1, Name: "系统管理", MenuType: menuTypeDir, Children: []*sys.SysMenu{
			{Id: 2, Name: "用户管理", MenuType: menuTypeMenu, Children: []*sys.SysMenu{
				{Id: 3, Name: "用户新增", MenuType: menuTypeButton, Perm: "sys:user:create"},
			}},
		}},
	}
	stripped := StripButtons(tree)
	sysDir := stripped[0]
	if got := idsOf(sysDir.GetChildren()); !reflect.DeepEqual(got, []int64{2}) {
		t.Fatalf("按钮剔除后系统管理应只剩用户管理，实际 %v", got)
	}
	if c := findChild(sysDir, "用户管理"); len(c.GetChildren()) != 0 {
		t.Fatalf("用户管理的按钮子级应被剔除，实际 %v", c.GetChildren())
	}
	//原树不被破坏（克隆语义，GetInfo 还要拿原树收集 perms）
	if c := findChild(tree[0], "用户管理"); len(c.GetChildren()) != 1 {
		t.Fatalf("StripButtons 应为克隆语义，原树不变")
	}
}

// TestBuildRouters 菜单树转前端路由：顶层绝对路径、目录下相对段、根"/"子节点拼全路径、按钮剔除、name 驼峰
func TestBuildRouters(t *testing.T) {
	tree := []*sys.SysMenu{
		{Id: 1, Name: "系统管理", Path: "/system", Icon: "Setting", MenuType: menuTypeDir,
			Children: []*sys.SysMenu{
				{Id: 2, Name: "用户管理", Path: "user", Component: "system/UserList", Icon: "User", MenuType: menuTypeMenu,
					Children: []*sys.SysMenu{
						{Id: 3, Name: "用户新增", MenuType: menuTypeButton, Perm: "sys:user:create"}, //应被剔除
					}},
			}},
		{Id: 4, Name: "仪表盘", Path: "/dashboard", Component: "Dashboard", Icon: "Odometer", MenuType: menuTypeMenu},
		{Id: 5, Name: "根目录", Path: "/", Icon: "Menu", MenuType: menuTypeDir,
			Children: []*sys.SysMenu{
				{Id: 6, Name: "独立页", Path: "/standalone", Component: "standalone/List", MenuType: menuTypeMenu}, //根"/"子节点拼全路径
			}},
		{Id: 7, Name: "相对顶层", Path: "relative", MenuType: menuTypeMenu}, //顶层缺 / 自动补
	}
	routers := BuildRouters(tree)
	if len(routers) != 4 {
		t.Fatalf("应有 4 条顶层路由（按钮剔除），实际 %d", len(routers))
	}

	sysRoute := routers[0]
	if sysRoute.GetName() != "System" || sysRoute.GetPath() != "/system" || sysRoute.GetComponent() != "" {
		t.Fatalf("目录路由 name/path/component = %s/%s/%s, want System///system",
			sysRoute.GetName(), sysRoute.GetPath(), sysRoute.GetComponent())
	}
	if len(sysRoute.GetChildren()) != 1 {
		t.Fatalf("系统管理应只剩 1 个子路由（按钮剔除），实际 %d", len(sysRoute.GetChildren()))
	}
	userRoute := sysRoute.GetChildren()[0]
	if userRoute.GetPath() != "user" || userRoute.GetComponent() != "system/UserList" || userRoute.GetName() != "SystemUser" {
		t.Fatalf("子菜单路由 path/name = %s/%s, want user/SystemUser（全路径驼峰）",
			userRoute.GetPath(), userRoute.GetName())
	}

	dash := routers[1]
	if dash.GetPath() != "/dashboard" || dash.GetComponent() != "Dashboard" || dash.GetName() != "Dashboard" {
		t.Fatalf("顶层菜单路由 = %s/%s/%s, want /dashboard/Dashboard/Dashboard",
			dash.GetName(), dash.GetPath(), dash.GetComponent())
	}

	rootDir := routers[2]
	if len(rootDir.GetChildren()) != 1 {
		t.Fatalf("根目录应含 1 个子路由，实际 %d", len(rootDir.GetChildren()))
	}
	standalone := rootDir.GetChildren()[0]
	if standalone.GetPath() != "/standalone" {
		t.Fatalf("根\"/\"目录子节点应拼全路径，实际 %s", standalone.GetPath())
	}

	relative := routers[3]
	if relative.GetPath() != "/relative" {
		t.Fatalf("顶层缺 / 应自动补全，实际 %s", relative.GetPath())
	}
}

// TestRouterName 路由 name 生成：分段驼峰、分隔符归一、空路径兜底
func TestRouterName(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"/system", "System"},
		{"/datasource/datasource", "DatasourceDatasource"},
		{"/user-manager", "UserManager"},
		{"/ai_chat/list", "AiChatList"},
		{"", "Index"},
		{"/", "Index"},
	}
	for _, c := range cases {
		if got := routerName(c.path); got != c.want {
			t.Fatalf("routerName(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

// TestHasCycle 环检测细分：正常链/孤儿/自环/互环
func TestHasCycle(t *testing.T) {
	a := menu(1, 2, "A", menuTypeDir, "")
	b := menu(2, 1, "B", menuTypeDir, "")
	byId := map[int64]*sys.SysMenu{1: a, 2: b}
	if !hasCycle(a, byId) || !hasCycle(b, byId) {
		t.Fatalf("互指节点应检出环")
	}
	orphan := menu(3, 999, "孤儿", menuTypeDir, "")
	if hasCycle(orphan, byId) {
		t.Fatalf("孤儿不是环")
	}
	self := menu(4, 4, "自环", menuTypeDir, "")
	byId[4] = self
	if !hasCycle(self, byId) {
		t.Fatalf("自环应检出")
	}
	root := menu(5, 0, "根", menuTypeDir, "")
	byId[5] = root
	if hasCycle(root, byId) {
		t.Fatalf("根节点无环")
	}
}

// TestJoinRoutePath 路径拼接边界：绝对子路径透传、空段、根父
func TestJoinRoutePath(t *testing.T) {
	cases := []struct {
		parent, child, want string
	}{
		{"", "user", "/user"},
		{"/", "user", "/user"},
		{"/system", "user", "/system/user"},
		{"/system", "/abs", "/abs"},
		{"/system", "", "/system"},
		{"/system", " spaced ", "/system/spaced"}, //相对段带空白：去空白后拼接
		{"/system", " /spaced ", "/spaced"},       //绝对子路径（含首尾空白）：去空白后透传
	}
	for _, c := range cases {
		if got := joinRoutePath(c.parent, c.child); got != c.want {
			t.Fatalf("joinRoutePath(%q, %q) = %q, want %q", c.parent, c.child, got, c.want)
		}
	}
}
