package logic

import "strings"

// mergeOwner 合并负责人字段（纯函数，便于单元测试）：
// 更新请求未带负责人（空/纯空白）时保留库中原值，避免表单漏传误清空
func mergeOwner(newOwner, oldOwner string) string {
	if strings.TrimSpace(newOwner) == "" {
		return oldOwner
	}
	return newOwner
}
