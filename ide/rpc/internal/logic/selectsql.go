package logic

import (
	"fmt"
	"strings"
)

// ValidateSelectSql 脚本执行前的安全闸（纯函数，便于单测），思路参考 tag 服务 ValidateRuleSql：
//  1. 剥掉前导注释（-- 行注释、# 行注释、块注释）与前导括号后，必须以 select 开头（只读限制，
//     杜绝 update/delete/drop 等写语句，也拒绝 with ... select 的 CTE 包装）；
//  2. 禁止分号（单语句限制，天然拒绝 `select 1; drop table ...` 等多语句注入片段，
//     结尾分号同样拒绝，与 tag 服务口径一致）；
//  3. 已知限制：字符串字面量中的分号同样会被拒绝（宁严勿漏）。
func ValidateSelectSql(sqlText string) error {
	s, ok := stripLeadingNoise(sqlText)
	if !ok {
		return fmt.Errorf("SQL 存在未闭合的块注释")
	}
	if s == "" {
		return fmt.Errorf("SQL 不能为空")
	}
	if len(s) < 6 || !strings.EqualFold(s[:6], "select") {
		return fmt.Errorf("仅允许 SELECT 查询")
	}
	//词边界校验：select 后必须是空白或结尾，杜绝 selectx 之类的标识符冒充
	if tail := s[6:]; tail != "" && !isSpaceByte(tail[0]) {
		return fmt.Errorf("仅允许 SELECT 查询")
	}
	if strings.Contains(s, ";") {
		return fmt.Errorf("仅允许单条语句，禁止分号（拒绝 ; drop 等多语句注入片段）")
	}
	return nil
}

// stripLeadingNoise 循环剥掉语句头部的空白、注释与前导括号，返回剩余语句文本；
// ok=false 表示出现未闭合的块注释。循环剥离支持注释与括号任意交错（如 "/* a */ ( /* b */ select ..."）
func stripLeadingNoise(sqlText string) (string, bool) {
	s := sqlText
	for {
		//跳过前导空白
		for len(s) > 0 && isSpaceByte(s[0]) {
			s = s[1:]
		}
		switch {
		case strings.HasPrefix(s, "--"), strings.HasPrefix(s, "#"):
			//行注释：跳到行尾（无行尾则整段只剩注释）
			idx := strings.IndexByte(s, '\n')
			if idx < 0 {
				return "", true
			}
			s = s[idx+1:]
		case strings.HasPrefix(s, "/*"):
			//块注释：未闭合视为非法
			idx := strings.Index(s, "*/")
			if idx < 0 {
				return "", false
			}
			s = s[idx+2:]
		case strings.HasPrefix(s, "("):
			//前导括号剥离，支持 (select ...) 包装
			s = s[1:]
		default:
			return s, true
		}
	}
}

// isSpaceByte 判断字节是否为空白字符
func isSpaceByte(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}
