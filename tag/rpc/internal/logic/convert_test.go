package logic

import (
	"testing"

	"datacenter/tag/model"
	"datacenter/tag/rpc/tag"
)

// TestObjectNameMap 对象名映射纯函数测试：按 id 建映射，空列表返回空映射
func TestObjectNameMap(t *testing.T) {
	objs := []*model.TagObject{
		{Id: 1, Name: "会员"},
		{Id: 2, Name: "商品"},
	}
	m := objectNameMap(objs)
	if len(m) != 2 || m[1] != "会员" || m[2] != "商品" {
		t.Fatalf("映射结果不符合预期: %v", m)
	}
	// 空列表
	if m := objectNameMap(nil); len(m) != 0 {
		t.Fatalf("空列表应返回空映射: %v", m)
	}
}

// TestWithObjectName 回填对象名纯函数测试：nil 安全
func TestWithObjectName(t *testing.T) {
	// 正常回填（每次用全新 def，避免上个用例回填值串场）
	got := withObjectName(&tag.TagDef{Id: 10}, &model.TagObject{Id: 1, Name: "会员"})
	if got.ObjectName != "会员" {
		t.Fatalf("期望回填对象名 会员，实际 %q", got.ObjectName)
	}
	// 对象为 nil 安全跳过
	if got := withObjectName(&tag.TagDef{Id: 10}, nil); got.ObjectName != "" {
		t.Fatalf("nil 对象不应回填，实际 %q", got.ObjectName)
	}
	// 定义为 nil 安全返回
	if got := withObjectName(nil, &model.TagObject{Id: 1, Name: "会员"}); got != nil {
		t.Fatalf("nil 定义应返回 nil")
	}
}
