package logic

import (
	"fmt"
	"strings"
)

// sqlguard.go AI 生成 SQL 的发送侧安全校验（纯函数，便于单测）。
// 与 ide-rpc 的 ValidateSelectSql 同一套口径（思路参考其实现风格），但独立自写：
// ide 的实现位于 internal 包无法跨服务复用。定位说明：
//   - 本校验在 LLM 生成侧提前拦截（Ask 审计、AskAndRun 执行前都会过这一道），
//     把"LLM 幻觉出写语句/多语句"挡在审计与执行之前；
//   - AskAndRun 执行侧另有二次防护：SQL 经 ide-rpc 安全闸（仅 SELECT、行数截断）真实执行，
//     即使本侧被绕过，写语句也到不了数据库。双侧独立，纵深防御。
//
// 规则（宁严勿漏）：
//  1. 剥掉前导空白/注释（-- # /* */）/前导括号后必须以 select 开头（只读限制，
//     也拒绝 with ... select 的 CTE 包装——演示场景用不到，从源头收紧）；
//  2. 全文禁止分号（单语句限制，天然拒绝 `select 1; drop table ...` 多语句注入片段，
//     结尾分号同样拒绝；已知限制：字符串字面量中的分号也会被拒，可接受）。
const generatedSqlMaxLen = 4096 //生成 SQL 长度上限（一条分析查询不该太长，防御异常超长输出）

// ValidateGeneratedSql 校验 LLM 生成的 SQL 是否为单条只读 SELECT；通过返回 nil，否则带原因
func ValidateGeneratedSql(sqlText string) error {
	s := strings.TrimSpace(sqlText)
	if s == "" {
		return fmt.Errorf("生成 SQL 为空")
	}
	if len(s) > generatedSqlMaxLen {
		return fmt.Errorf("生成 SQL 超长(%d 字节,上限 %d)", len(s), generatedSqlMaxLen)
	}

	//1. 剥掉前导噪音（注释/括号/空白）后必须以 select 开头
	s, ok := stripLeadingNoise(s)
	if !ok {
		return fmt.Errorf("生成 SQL 存在未闭合的块注释")
	}
	if s == "" {
		return fmt.Errorf("仅允许 SELECT 查询")
	}
	if len(s) < 6 || !strings.EqualFold(s[:6], "select") {
		return fmt.Errorf("仅允许 SELECT 查询")
	}
	//词边界校验：select 后必须是空白或结尾，杜绝 selectx 之类的标识符冒充
	if tail := s[6:]; tail != "" && !isSpaceByte(tail[0]) {
		return fmt.Errorf("仅允许 SELECT 查询")
	}

	//2. 单语句限制：全文禁止分号（拒绝 ; drop 等多语句注入片段）
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
