package model

import (
	"regexp"
	"testing"
)

// permPatternTest 与 logic 层保存校验一致的 perm 格式（段允许驼峰，对齐 RuoYi resetPwd 惯例）
var permPatternTest = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(?::[a-zA-Z][a-zA-Z0-9_]*)+$`)

// TestMenuSeedsIntegrity 种子树完整性：兄弟不重名（父+名幂等定位的前提）、
// perm 全局唯一且格式合法、父子类型匹配（目录←目录；菜单←目录；按钮←菜单）
func TestMenuSeedsIntegrity(t *testing.T) {
	perms := make(map[string]string) //perm → 菜单名（查重/报错定位）

	var walk func(nodes []seedMenu, parentType int64, path []string)
	walk = func(nodes []seedMenu, parentType int64, path []string) {
		seenNames := make(map[string]struct{}, len(nodes))
		for _, s := range nodes {
			key := joinPath(path, s.Name)
			// 兄弟重名会破坏 FindOneByName(父+名) 幂等定位
			if _, ok := seenNames[s.Name]; ok {
				t.Fatalf("同级菜单重名：%s", key)
			}
			seenNames[s.Name] = struct{}{}
			// 菜单类型白名单
			switch s.MenuType {
			case MenuTypeDir, MenuTypeMenu, MenuTypeButton:
			default:
				t.Fatalf("菜单类型非法：%s type=%d", key, s.MenuType)
			}
			// 父子类型匹配（根节点 parentType=0）
			if parentType != 0 {
				switch s.MenuType {
				case MenuTypeDir, MenuTypeMenu:
					if parentType != MenuTypeDir {
						t.Fatalf("目录/菜单父必须为目录：%s", key)
					}
				case MenuTypeButton:
					if parentType != MenuTypeMenu {
						t.Fatalf("按钮父必须为菜单：%s", key)
					}
				}
			} else if s.MenuType == MenuTypeButton {
				t.Fatalf("顶层不允许按钮：%s", key)
			}
			// perm 格式与全局唯一
			if s.MenuType == MenuTypeButton {
				if !permPatternTest.MatchString(s.Perm) {
					t.Fatalf("按钮 perm 格式非法：%s perm=%q", key, s.Perm)
				}
				if owner, ok := perms[s.Perm]; ok {
					t.Fatalf("perm 全局重复 %q：%s 与 %s", s.Perm, owner, key)
				}
				perms[s.Perm] = key
			} else if s.Perm != "" {
				t.Fatalf("目录/菜单不应配 perm：%s perm=%q", key, s.Perm)
			}
			walk(s.Children, s.MenuType, append(path, s.Name))
		}
	}
	walk(menuSeeds(), 0, nil)

	// 种子规模底线：perm 数量（系统管理 13 + 数据中台 13*4 + 扩展 6 = 71）防种子被误删
	if len(perms) < 60 {
		t.Fatalf("种子按钮权限数量异常：仅 %d 项", len(perms))
	}
}

func joinPath(path []string, name string) string {
	out := ""
	for _, p := range path {
		out += p + "/"
	}
	return out + name
}

// TestSeedMenuCounts 种子结构底线：系统管理 3 页、数据中台 13 组模块齐全
func TestSeedMenuCounts(t *testing.T) {
	seeds := menuSeeds()
	if len(seeds) < 15 {
		t.Fatalf("顶层种子应含 系统管理+仪表盘+13 组数据中台模块，实际 %d", len(seeds))
	}
	names := make(map[string]struct{}, len(seeds))
	for _, s := range seeds {
		names[s.Name] = struct{}{}
	}
	for _, want := range []string{"系统管理", "仪表盘", "数据源", "元数据", "数据服务", "资产目录",
		"数据质量", "指标管理", "标签管理", "数据标准", "主数据", "调度管理", "数据IDE", "AI 问数", "数据运维"} {
		if _, ok := names[want]; !ok {
			t.Fatalf("种子缺少顶层模块：%s", want)
		}
	}
}
