package logic

import (
	"strings"
	"testing"
)

// 测试生成 SQL 安全校验（发送侧）：仅放行单条 SELECT；
// 注意 AskAndRun 执行侧 ide 安全闸还有二次防护，本校验是第一道（见 sqlguard.go 注释）
func TestValidateGeneratedSql(t *testing.T) {
	cases := []struct {
		name    string
		sqlText string
		wantErr bool
	}{
		//放行：常规分析查询形态
		{name: "简单查询", sqlText: "SELECT COUNT(*) FROM dm_orders", wantErr: false},
		{name: "小写查询", sqlText: "select sum(amount) from dm_orders limit 100", wantErr: false},
		{name: "前导空白换行", sqlText: "\n  SELECT 1", wantErr: false},
		{name: "前导行注释", sqlText: "-- 分析查询\nselect 1", wantErr: false},
		{name: "前导块注释", sqlText: "/* 生成 */ select 1", wantErr: false},
		{name: "前导括号包装", sqlText: "(select 1)", wantErr: false},
		{name: "注释与括号交错", sqlText: "/* a */ ( -- b\n select 1", wantErr: false},

		//拒绝：写语句/非 SELECT 开头
		{name: "delete", sqlText: "delete from dm_orders", wantErr: true},
		{name: "update", sqlText: "update dm_orders set amount=0", wantErr: true},
		{name: "drop", sqlText: "drop table dm_orders", wantErr: true},
		{name: "insert", sqlText: "insert into t values(1)", wantErr: true},
		{name: "空串", sqlText: "", wantErr: true},
		{name: "纯空白", sqlText: "   \n\t", wantErr: true},
		{name: "selectx 标识符冒充", sqlText: "selectx from t", wantErr: true},
		{name: "只剩注释", sqlText: "-- nothing", wantErr: true},
		{name: "未闭合块注释", sqlText: "/* select 1", wantErr: true},
		{name: "CTE 包装拒绝", sqlText: "with t as (select 1) select * from t", wantErr: true},

		//拒绝：多语句/分号注入片段（宁严勿漏，字面量分号同样拒绝）
		{name: "多语句注入", sqlText: "select 1; drop table dm_orders", wantErr: true},
		{name: "结尾分号", sqlText: "select 1;", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateGeneratedSql(c.sqlText)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidateGeneratedSql(%q) err = %v, wantErr = %v", c.sqlText, err, c.wantErr)
			}
		})
	}
}

// 测试超长 SQL 拒绝：防御异常超长 LLM 输出
func TestValidateGeneratedSqlTooLong(t *testing.T) {
	long := "select " + strings.Repeat("1,", generatedSqlMaxLen) + "1"
	if err := ValidateGeneratedSql(long); err == nil || !strings.Contains(err.Error(), "超长") {
		t.Fatalf("超长 SQL err = %v, 期望带超长原因", err)
	}
}
