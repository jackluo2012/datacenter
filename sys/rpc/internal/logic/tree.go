package logic

import (
	"strings"

	"datacenter/sys/rpc/sys"
)

// 菜单类型（对齐 sys_menu.menu_type / model 层常量，logic 内独立声明避免包间耦合方向混乱）
const (
	menuTypeDir    = 1 // 目录
	menuTypeMenu   = 2 // 菜单
	menuTypeButton = 3 // 按钮
)

// BuildMenuTree 全量内存建树（纯函数，不依赖 DB）：
//   - ParentId=0 挂根；父缺失（孤儿）或父链成环（含自环）的节点兜底挂根，保证所有节点可达、不丢失
//   - 同层顺序保持入参顺序（调用方按 sort,id 排序后传入）
//   - 入参节点 Children 会被重建，调用方不应复用其原 children
func BuildMenuTree(menus []*sys.SysMenu) []*sys.SysMenu {
	byId := make(map[int64]*sys.SysMenu, len(menus))
	for _, m := range menus {
		m.Children = nil
		byId[m.GetId()] = m
	}
	roots := make([]*sys.SysMenu, 0, len(menus))
	for _, m := range menus {
		parentId := m.GetParentId()
		if parentId == 0 {
			roots = append(roots, m)
			continue
		}
		parent, ok := byId[parentId]
		// 孤儿（父不存在）或环节点：兜底挂根，避免节点丢失
		if !ok || hasCycle(m, byId) {
			roots = append(roots, m)
			continue
		}
		parent.Children = append(parent.Children, m)
	}
	return roots
}

// hasCycle 沿父链上溯检测环：回到已访问节点说明成环（含自环 ParentId==Id）；
// 父缺失（孤儿）不算环，由调用方单独兜底
func hasCycle(m *sys.SysMenu, byId map[int64]*sys.SysMenu) bool {
	seen := make(map[int64]struct{}, 8)
	cur := m
	for {
		if cur.GetParentId() == 0 {
			return false
		}
		if _, ok := seen[cur.GetId()]; ok {
			return true
		}
		seen[cur.GetId()] = struct{}{}
		parent, ok := byId[cur.GetParentId()]
		if !ok {
			return false
		}
		cur = parent
	}
}

// StripButtons 递归剔除按钮节点（GetInfo/GetRouters 菜单树仅保留目录+菜单）。
// 克隆语义：返回新节点（逐字段拷贝，不整结构体复制——pb 结构含 protoimpl 状态锁），
// 入参树不被修改，GetInfo 可先 CollectPerms 再 StripButtons 复用同一棵树
func StripButtons(menus []*sys.SysMenu) []*sys.SysMenu {
	out := make([]*sys.SysMenu, 0, len(menus))
	for _, m := range menus {
		if m.GetMenuType() == menuTypeButton {
			continue
		}
		out = append(out, &sys.SysMenu{
			Id:        m.GetId(),
			ParentId:  m.GetParentId(),
			Name:      m.GetName(),
			Path:      m.GetPath(),
			Component: m.GetComponent(),
			Sort:      m.GetSort(),
			Icon:      m.GetIcon(),
			MenuType:  m.GetMenuType(),
			Perm:      m.GetPerm(),
			Visible:   m.GetVisible(),
			Beid:      m.GetBeid(),
			Children:  StripButtons(m.GetChildren()),
		})
	}
	return out
}

// CollectPerms 收集按钮权限标识集合（menu_type=按钮 且 perm 非空；去重保序）
func CollectPerms(menus []*sys.SysMenu) []string {
	seen := make(map[string]struct{})
	perms := make([]string, 0)
	collectPerms(menus, seen, &perms)
	return perms
}

// collectPerms CollectPerms 递归实现（seen 去重，perms 指针累积）
func collectPerms(menus []*sys.SysMenu, seen map[string]struct{}, perms *[]string) {
	for _, m := range menus {
		if perm := m.GetPerm(); perm != "" {
			if _, ok := seen[perm]; !ok {
				seen[perm] = struct{}{}
				*perms = append(*perms, perm)
			}
		}
		collectPerms(m.GetChildren(), seen, perms)
	}
}

// BuildRouters 用户菜单树转前端动态路由（RuoYi 约定，纯函数）：
//   - 目录：path 顶层保持绝对路径（非 / 开头补 /），component 留空由前端按布局渲染
//   - 菜单：目录下用相对路径段；目录 path 为 "/" 时子节点拼全路径
//   - component 透传，前端按 web/src/views/<component>.vue 解析（如 datasource/DataSourceList
//     → web/src/views/datasource/DataSourceList.vue）
//   - name 由全路径分段驼峰生成，保证唯一；按钮节点剔除
func BuildRouters(menus []*sys.SysMenu) []*sys.RouterBranch {
	return routersOf(StripButtons(menus), "")
}

// routersOf BuildRouters 递归实现（parentPath 为父节点全路径，顶层传空串）
func routersOf(menus []*sys.SysMenu, parentPath string) []*sys.RouterBranch {
	out := make([]*sys.RouterBranch, 0, len(menus))
	for _, m := range menus {
		if m.GetMenuType() == menuTypeButton {
			continue
		}
		full := joinRoutePath(parentPath, m.GetPath())
		branch := &sys.RouterBranch{
			Name:      routerName(full),
			Path:      branchPath(m.GetPath(), parentPath),
			Component: m.GetComponent(),
			Icon:      m.GetIcon(),
			Children:  routersOf(m.GetChildren(), full),
		}
		out = append(out, branch)
	}
	return out
}

// branchPath 计算路由 path：
//   - 顶层节点（parentPath 为空）：绝对路径（补 / 前缀）
//   - 父目录 path 为 "/"：子节点拼全路径（任务约定，避免相对根歧义）
//   - 其余目录下子节点：相对路径段原样透传（含用户自配的绝对路径）
func branchPath(path, parentPath string) string {
	if parentPath == "" {
		return ensureAbs(path)
	}
	if parentPath == "/" {
		return joinRoutePath("/", path)
	}
	return path
}

// joinRoutePath 拼接父子路由路径：子为绝对路径（/ 开头）时直接返回，
// 父为空或根时补 / 前缀，否则 parent/child
func joinRoutePath(parent, child string) string {
	child = strings.TrimSpace(child)
	if child == "" {
		return parent
	}
	if strings.HasPrefix(child, "/") {
		return child
	}
	if parent == "" || parent == "/" {
		return "/" + child
	}
	return parent + "/" + child
}

// ensureAbs 顶层路由 path 保证以 / 开头
func ensureAbs(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}

// routerName 路由 name：按路径段转驼峰拼接（/system → System，
// /datasource/datasource → DatasourceDatasource），分段分隔符为 / - _，保证 name 唯一
func routerName(path string) string {
	segs := strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '-' || r == '_'
	})
	var b strings.Builder
	for _, seg := range segs {
		b.WriteString(camelWord(seg))
	}
	if b.Len() == 0 {
		return "Index"
	}
	return b.String()
}

// camelWord 单词首字母大写（utf-8 安全，rest 保持原样）
func camelWord(word string) string {
	if word == "" {
		return ""
	}
	r := []rune(word)
	up := strings.ToUpper(string(r[0]))
	return up + string(r[1:])
}
