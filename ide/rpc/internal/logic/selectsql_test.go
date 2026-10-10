package logic

import (
	"strings"
	"testing"
)

// TestValidateSelectSql 执行脚本安全闸：仅放行单条 SELECT；
// 非 select 语句、多语句分号、注入片段必须拒绝；前导注释与括号包装放行
func TestValidateSelectSql(t *testing.T) {
	cases := []struct {
		name    string
		sqlText string
		wantErr bool
	}{
		{"合法 select", "select id, name from ods.t_user where id = 1", false},
		{"大写 SELECT", "SELECT 1", false},
		{"换行多段", "select id\nfrom t_user\nwhere status = 1", false},
		{"括号包装", "(select 1)", false},
		{"前导行注释+括号", "-- 查询用户\n(select id from t_user limit 10)", false},
		{"前导块注释", "/* 只读查询 */ select 1", false},
		{"前导井号注释", "# 查询\nselect 1", false},
		{"注释与括号交错", "/* a */ ( /* b */ select 1 )", false},
		{"空语句", "   ", true},
		{"纯注释", "-- 只有注释", true},
		{"非 select: update", "update t_user set name = 'x'", true},
		{"非 select: delete", "delete from t_user", true},
		{"非 select: insert", "insert into t(id) values (1)", true},
		{"非 select: drop", "drop table t_user", true},
		{"非 select: show", "show tables", true},
		{"cte 包装拒绝", "with c as (select 1) select * from c", true},
		{"标识符冒充 select", "selectx 1", true},
		{"多语句", "select 1; select 2", true},
		{"注入片段 ; drop", "select 1; drop table t", true},
		{"结尾分号也拒绝", "select 1;", true},
		{"块注释未闭合", "/* 未闭合 select 1", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateSelectSql(c.sqlText)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidateSelectSql(%q) err = %v, wantErr = %v", c.sqlText, err, c.wantErr)
			}
		})
	}
}

// TestValidateSelectSqlErrorText 多语句注入片段拒绝时的错误信息应说明分号限制
func TestValidateSelectSqlErrorText(t *testing.T) {
	err := ValidateSelectSql("select 1; drop table t")
	if err == nil {
		t.Fatal("select 1; drop 必须被拒绝")
	}
	if !strings.Contains(err.Error(), "分号") {
		t.Fatalf("错误信息应提示分号限制, got: %v", err)
	}
	//非 select 拒绝时的错误信息应说明仅允许 SELECT
	err = ValidateSelectSql("delete from t_user")
	if err == nil {
		t.Fatal("delete 必须被拒绝")
	}
	if !strings.Contains(err.Error(), "SELECT") {
		t.Fatalf("错误信息应提示仅允许 SELECT, got: %v", err)
	}
}
