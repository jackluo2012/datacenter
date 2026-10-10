package logic

import (
	"strings"
	"testing"
)

// TestValidateRuleSql 规则 SQL 校验：非 select / 缺 as object_id / 缺 as tag_value /
// 注入片段 `; drop` 等必须拒绝，合法规则放行
func TestValidateRuleSql(t *testing.T) {
	valid := "select id as object_id, case when score >= 90 then 'A' else 'B' end as tag_value from ods.t_user where status = 1"
	cases := []struct {
		name    string
		sqlText string
		wantErr bool
	}{
		{"合法规则", valid, false},
		{"别名大小写不敏感", "SELECT id AS OBJECT_ID, 'x' AS TAG_VALUE FROM t", false},
		{"多余空白容忍", "  select id as object_id, level as tag_value from t  ", false},
		{"空语句", "   ", true},
		{"过短语句", "sele", true},
		{"非 select: update", "update t_user set level = 1", true},
		{"非 select: delete", "delete from t_user", true},
		{"非 select: insert", "insert into t(id) values (1)", true},
		{"缺 as object_id", "select id, 'x' as tag_value from t", true},
		{"缺 as tag_value", "select id as object_id from t", true},
		{"两个别名都缺", "select id, name from t", true},
		{"冒充别名不带词边界", "select id as object_id2, 'x' as tag_value2 from t", true},
		{"注入片段 ; drop", "select id as object_id, 'x' as tag_value from t; drop table t", true},
		{"注入片段 ; delete", "select id as object_id, 'x' as tag_value from t;delete from dc_tag_value", true},
		{"结尾分号也拒绝", "select id as object_id, 'x' as tag_value from t;", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateRuleSql(c.sqlText)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidateRuleSql(%q) err = %v, wantErr = %v", c.sqlText, err, c.wantErr)
			}
		})
	}
}

// TestValidateRuleSqlErrorText 注入片段拒绝时的错误信息应说明分号限制
func TestValidateRuleSqlErrorText(t *testing.T) {
	err := ValidateRuleSql("select id as object_id, 'x' as tag_value from t; drop table t")
	if err == nil {
		t.Fatal("注入片段 ; drop 必须被拒绝")
	}
	if !strings.Contains(err.Error(), "分号") {
		t.Fatalf("错误信息应提示分号限制, got: %v", err)
	}
}

// TestValidTagType tag_type 白名单：仅 rule/fact/model
func TestValidTagType(t *testing.T) {
	for _, typ := range []string{"rule", "fact", "model"} {
		if !ValidTagType(typ) {
			t.Fatalf("tag_type %q 应在白名单内", typ)
		}
	}
	for _, typ := range []string{"", "sql", "RULE", "model2", " rule"} {
		if ValidTagType(typ) {
			t.Fatalf("tag_type %q 不应在白名单内", typ)
		}
	}
}

// TestValidObjectType object_type 白名单：仅 user/member/goods
func TestValidObjectType(t *testing.T) {
	for _, typ := range []string{"user", "member", "goods"} {
		if !ValidObjectType(typ) {
			t.Fatalf("object_type %q 应在白名单内", typ)
		}
	}
	for _, typ := range []string{"", "order", "User", "shop"} {
		if ValidObjectType(typ) {
			t.Fatalf("object_type %q 不应在白名单内", typ)
		}
	}
}
