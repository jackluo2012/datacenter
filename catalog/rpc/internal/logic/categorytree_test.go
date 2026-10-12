package logic

import (
	"testing"

	"datacenter/catalog/model"
)

// mkCategory 构造测试用分类行
func mkCategory(id, parentId int64) *model.AssetCategory {
	return &model.AssetCategory{Id: id, ParentId: parentId}
}

// TestInSubtree 成环校验纯函数测试：
// target 落在以 root 为根的子树内（含 root 自身）返回 true；旁支与空树返回 false
func TestInSubtree(t *testing.T) {
	// 树结构: 1(root) → 2 → 3；旁支: 1 → 9
	tree := []*model.AssetCategory{
		mkCategory(1, 0),
		mkCategory(2, 1),
		mkCategory(3, 2),
		mkCategory(9, 1),
	}
	cases := []struct {
		name   string
		root   int64
		target int64
		want   bool
	}{
		{name: "目标是自身", root: 2, target: 2, want: true},
		{name: "目标是直接子分类", root: 1, target: 2, want: true},
		{name: "目标是孙分类", root: 1, target: 3, want: true},
		{name: "目标是旁支", root: 2, target: 9, want: false},
		{name: "目标是父分类", root: 2, target: 1, want: false},
		{name: "目标不在树中", root: 1, target: 99, want: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := inSubtree(tree, c.root, c.target); got != c.want {
				t.Fatalf("inSubtree(root=%d, target=%d) 期望 %v，实际 %v", c.root, c.target, c.want, got)
			}
		})
	}
}

// TestInSubtreeEmpty 空分类集合只可能命中 root 自身
func TestInSubtreeEmpty(t *testing.T) {
	if !inSubtree(nil, 5, 5) {
		t.Fatal("root 自身应命中")
	}
	if inSubtree(nil, 5, 6) {
		t.Fatal("空集合不应命中其他节点")
	}
}
