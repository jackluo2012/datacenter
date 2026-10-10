package logic

import (
	"fmt"
	"regexp"
	"strings"
)

// 标签类型
const (
	TagTypeRule  = "rule"  //规则标签（rule_sql 跑目标库产出）
	TagTypeFact  = "fact"  //事实标签
	TagTypeModel = "model" //模型标签
)

// 标签对象类型
const (
	ObjectTypeUser   = "user"
	ObjectTypeMember = "member"
	ObjectTypeGoods  = "goods"
)

// defaultPkColumn 对象未显式指定主键列时的默认值
const defaultPkColumn = "id"

var (
	// ruleAliasObjectIdRe 规则 SQL 结果集必须包含 as object_id 别名列（忽略大小写、词边界匹配，
	// "as object_id2" 这类冒充别名无法通过）
	ruleAliasObjectIdRe = regexp.MustCompile(`(?i)\bas\s+object_id\b`)
	// ruleAliasTagValueRe 规则 SQL 结果集必须包含 as tag_value 别名列
	ruleAliasTagValueRe = regexp.MustCompile(`(?i)\bas\s+tag_value\b`)
)

// ValidTagType 标签类型白名单校验（rule/fact/model）
func ValidTagType(typ string) bool {
	switch typ {
	case TagTypeRule, TagTypeFact, TagTypeModel:
		return true
	default:
		return false
	}
}

// ValidObjectType 标签对象类型白名单校验（user/member/goods）
func ValidObjectType(typ string) bool {
	switch typ {
	case ObjectTypeUser, ObjectTypeMember, ObjectTypeGoods:
		return true
	default:
		return false
	}
}

// ValidateRuleSql 规则 SQL 校验（纯函数，便于单测）：
// 1) 非空；2) 必须以 select 开头（只读限制，杜绝 update/delete 等写语句）；
// 3) 必须包含 as object_id 与 as tag_value 两列别名（打标执行按列名取数）；
// 4) 禁止分号（单语句限制，天然拒绝 `; drop table ...` 等注入片段）。
func ValidateRuleSql(sqlText string) error {
	s := strings.TrimSpace(sqlText)
	if s == "" {
		return fmt.Errorf("rule_sql 不能为空")
	}
	if len(s) < 6 || !strings.EqualFold(s[:6], "select") {
		return fmt.Errorf("rule_sql 必须以 select 开头")
	}
	if strings.Contains(s, ";") {
		return fmt.Errorf("rule_sql 仅允许单条语句，禁止分号（拒绝 ; drop 等注入片段）")
	}
	if !ruleAliasObjectIdRe.MatchString(s) {
		return fmt.Errorf("rule_sql 缺少 as object_id 别名列")
	}
	if !ruleAliasTagValueRe.MatchString(s) {
		return fmt.Errorf("rule_sql 缺少 as tag_value 别名列")
	}
	return nil
}
