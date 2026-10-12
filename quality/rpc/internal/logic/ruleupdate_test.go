package logic

import "testing"

// TestMergeOwner 负责人合并纯函数测试：新值空/纯空白保留原值，否则取新值
func TestMergeOwner(t *testing.T) {
	cases := []struct {
		name     string
		newOwner string
		oldOwner string
		want     string
	}{
		{name: "新值非空取新值", newOwner: "张三", oldOwner: "李四", want: "张三"},
		{name: "新值为空保留原值", newOwner: "", oldOwner: "李四", want: "李四"},
		{name: "新值纯空白保留原值", newOwner: "   ", oldOwner: "李四", want: "李四"},
		{name: "新旧均空", newOwner: "", oldOwner: "", want: ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := mergeOwner(c.newOwner, c.oldOwner); got != c.want {
				t.Fatalf("mergeOwner(%q, %q) 期望 %q，实际 %q", c.newOwner, c.oldOwner, c.want, got)
			}
		})
	}
}
