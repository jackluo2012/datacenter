package logic

import (
	"errors"
	"regexp"
	"testing"
)

// TestValidateCircleRuleSql 圈选 SQL 校验：非 select / 缺 as object_id /
// 注入片段 `; drop` 等必须拒绝；合法圈选（含不带成员值列的纯圈人）放行
func TestValidateCircleRuleSql(t *testing.T) {
	// 合法：纯圈人，无 as tag_value 成员值列（列可选是圈选与打标规则 SQL 的核心差异）
	circleOnly := "select u.id as object_id from ods.t_user u join dc_tag_value v on v.object_pk = u.id where v.tag_key = 'vip' and v.tag_value = '高价值'"
	// 合法：带 as tag_value 成员值列
	withValue := "select id as object_id, level as tag_value from dm.t_member where level >= 3"
	cases := []struct {
		name    string
		sqlText string
		wantErr bool
	}{
		{"合法纯圈人无值列", circleOnly, false},
		{"合法带成员值列", withValue, false},
		{"别名大小写不敏感", "SELECT id AS OBJECT_ID FROM t", false},
		{"多余空白容忍", "  select id as object_id from t  ", false},
		{"空语句", "   ", true},
		{"过短语句", "sele", true},
		{"非 select: update", "update t_user set level = 1", true},
		{"非 select: delete", "delete from t_user", true},
		{"非 select: insert", "insert into t(id) values (1)", true},
		{"非 select: create", "create table t(id int)", true},
		{"缺 as object_id 别名", "select id, name from t", true},
		{"冒充别名不带词边界", "select id as object_id2 from t", true},
		{"注入片段 ; drop", "select id as object_id from t; drop table t", true},
		{"注入片段 ; delete", "select id as object_id from t;delete from dc_tag_value", true},
		{"结尾分号也拒绝", "select id as object_id from t;", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateCircleRuleSql(c.sqlText)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidateCircleRuleSql(%q) err = %v, wantErr = %v", c.sqlText, err, c.wantErr)
			}
		})
	}
}

// TestValidateCircleRuleSqlErrorText 注入片段拒绝时的错误信息应说明分号限制
func TestValidateCircleRuleSqlErrorText(t *testing.T) {
	err := ValidateCircleRuleSql("select id as object_id from t; drop table t")
	if err == nil {
		t.Fatal("注入片段 ; drop 必须被拒绝")
	}
	if !regexp.MustCompile("分号").MatchString(err.Error()) {
		t.Fatalf("错误信息应提示分号限制, got: %v", err)
	}
	// 圈选校验不应误用打标规则的 tag_value 必填逻辑（成员值列可选）
	err = ValidateCircleRuleSql("select id as object_id from t where status = 1")
	if err != nil {
		t.Fatalf("无成员值列的纯圈人不应被拒绝, got: %v", err)
	}
}

// TestGroupMaterializedTable 物化表名构造：仅接受正整数 group_id，
// 0/负数拒绝（int64 类型化入参 + 正整数校验，杜绝表名注入）；
// 合法值产出 dc_tag_group_member_<id>，且整体只含字母数字下划线安全字符
func TestGroupMaterializedTable(t *testing.T) {
	name, err := GroupMaterializedTable(42)
	if err != nil {
		t.Fatalf("group_id=42 应合法, got err: %v", err)
	}
	if name != "dc_tag_group_member_42" {
		t.Fatalf("物化表名 = %q, want %q", name, "dc_tag_group_member_42")
	}

	for _, gid := range []int64{0, -1, -42} {
		name, err := GroupMaterializedTable(gid)
		if err == nil {
			t.Fatalf("group_id=%d 必须被拒绝", gid)
		}
		if !errors.Is(err, ErrInvalidGroupId) {
			t.Fatalf("group_id=%d 应返回 ErrInvalidGroupId, got: %v", gid, err)
		}
		if name != "" {
			t.Fatalf("group_id=%d 被拒绝时表名应为空, got: %q", gid, name)
		}
	}

	// 合法表名整体必须匹配安全字符白名单（防拼接引入额外符号）
	safeRe := regexp.MustCompile(`^dc_tag_group_member_[1-9][0-9]*$`)
	for _, gid := range []int64{1, 8, 100000} {
		name, err := GroupMaterializedTable(gid)
		if err != nil {
			t.Fatalf("group_id=%d 应合法, got err: %v", gid, err)
		}
		if !safeRe.MatchString(name) {
			t.Fatalf("物化表名 %q 不符合安全字符白名单", name)
		}
	}
}

// TestIsMysqlTableNotExist 表不存在错误识别：1146 错误码 / doesn't exist 文本命中，
// 其他错误不误判（用于从未圈选时物化表缺失的优雅降级）
func TestIsMysqlTableNotExist(t *testing.T) {
	if !isMysqlTableNotExist(errors.New("Error 1146 (42S02): Table 'dm.dc_tag_group_member_1' doesn't exist")) {
		t.Fatal("1146 表不存在错误应被识别")
	}
	if !isMysqlTableNotExist(errors.New("table `dm.t` doesn't exist")) {
		t.Fatal("doesn't exist 文本应被识别")
	}
	for _, msg := range []string{"", "connection refused", "Error 1064: syntax error"} {
		if isMysqlTableNotExist(errors.New(msg)) {
			t.Fatalf("错误 %q 不应被识别为表不存在", msg)
		}
	}
	if isMysqlTableNotExist(nil) {
		t.Fatal("nil 错误不应被识别为表不存在")
	}
}
