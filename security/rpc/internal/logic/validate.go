package logic

import (
	"regexp"
	"strings"
)

// levelCodeRe 分级编码白名单：L1-L9
// （内置 L1 公开/L2 内部/L3 敏感/L4 机密，L5-L9 留给租户自定义级别）
var levelCodeRe = regexp.MustCompile(`^L[1-9]$`)

// ValidLevelCode 分级编码是否通过白名单（纯函数）
func ValidLevelCode(code string) bool {
	return levelCodeRe.MatchString(code)
}

// subjectTypeSet 授权主体类型白名单（subject = type:id）
var subjectTypeSet = map[string]struct{}{
	"user": {}, //平台用户
	"app":  {}, //应用/服务账号
	"role": {}, //角色组
}

// ValidSubject 授权主体格式校验（纯函数）：
// 需为 type:id 且 type ∈ user/app/role，id 段非空白（id 本身不限定字符集）
func ValidSubject(subject string) bool {
	parts := strings.SplitN(subject, ":", 2)
	if len(parts) != 2 {
		return false
	}
	if _, ok := subjectTypeSet[parts[0]]; !ok {
		return false
	}
	return strings.TrimSpace(parts[1]) != ""
}
