package model

import "testing"

// 临时探针：验证 menuSeeds 的类型分布（定位 acc.view 计数问题后删除）
func TestMenuSeedsTypeDist(t *testing.T) {
	seeds := menuSeeds()
	var all, view, btn int
	var walk func(ss []seedMenu)
	walk = func(ss []seedMenu) {
		for _, s := range ss {
			all++
			if s.MenuType == MenuTypeButton {
				btn++
			} else {
				view++
			}
			walk(s.Children)
		}
	}
	walk(seeds)
	t.Logf("all=%d view=%d button=%d", all, view, btn)
}
