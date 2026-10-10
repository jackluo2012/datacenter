package logic

import (
	"strings"
	"testing"
)

// TestSqlTypeOf 类型映射表：string→varchar(255)/int→bigint/decimal→decimal(18,4)/
// bool→tinyint(1)/date→date；白名单外返回空串
func TestSqlTypeOf(t *testing.T) {
	cases := []struct {
		fieldType string
		want      string
	}{
		{FieldTypeString, "varchar(255)"},
		{FieldTypeInt, "bigint"},
		{FieldTypeDecimal, "decimal(18,4)"},
		{FieldTypeBool, "tinyint(1)"},
		{FieldTypeDate, "date"},
		{"float", ""}, //白名单外
		{"INT", ""},   //大小写敏感
		{"", ""},      //空类型
	}
	for _, c := range cases {
		if got := SqlTypeOf(c.fieldType); got != c.want {
			t.Fatalf("SqlTypeOf(%q) = %q, want %q", c.fieldType, got, c.want)
		}
		if c.want == "" && ValidFieldType(c.fieldType) {
			t.Fatalf("类型 %q 不应在白名单", c.fieldType)
		}
		if c.want != "" && !ValidFieldType(c.fieldType) {
			t.Fatalf("类型 %q 应在白名单", c.fieldType)
		}
	}
}

// TestBuildCreateTableDDL 自动建表 DDL：内置列 biz_code/created_at + 字段列类型映射 + uk_biz
func TestBuildCreateTableDDL(t *testing.T) {
	fields, err := ValidateFieldsJson(`[{"name":"name","type":"string"},{"name":"age","type":"int"},{"name":"vip","type":"bool"}]`)
	if err != nil {
		t.Fatalf("模型定义非法: %v", err)
	}
	ddl := buildCreateTableDDL("`dc_mdm_data_customer`", fields)
	for _, want := range []string{
		"create table if not exists `dc_mdm_data_customer`",
		"`biz_code` varchar(128) NOT NULL PRIMARY KEY",
		"`name` varchar(255) DEFAULT NULL",
		"`age` bigint DEFAULT NULL",
		"`vip` tinyint(1) DEFAULT NULL",
		"`created_at` datetime(3)",
		"UNIQUE KEY `uk_biz` (`biz_code`)",
	} {
		if !strings.Contains(ddl, want) {
			t.Fatalf("DDL 缺少片段 %q:\n%s", want, ddl)
		}
	}
}

// TestBuildReplaceSQL replace into 语句拼装：biz_code + 字段列 + created_at，占位符个数对齐
func TestBuildReplaceSQL(t *testing.T) {
	fields, err := ValidateFieldsJson(`[{"name":"name","type":"string"},{"name":"age","type":"int"}]`)
	if err != nil {
		t.Fatalf("模型定义非法: %v", err)
	}
	sqlText, n := buildReplaceSQL("`db`.`tbl`", fields)
	want := "replace into `db`.`tbl` (`biz_code`,`name`,`age`,`created_at`) values (?,?,?,?)"
	if sqlText != want {
		t.Fatalf("replace SQL = %s, want %s", sqlText, want)
	}
	if n != 4 {
		t.Fatalf("占位符个数 = %d, want 4", n)
	}
}

// TestBuildTargetTable 限定名拼装：库名非空时 `db`.`table`，为空时 `table`
func TestBuildTargetTable(t *testing.T) {
	if got := buildTargetTable("crm", "dc_cus"); got != "`crm`.`dc_cus`" {
		t.Fatalf("buildTargetTable = %q", got)
	}
	if got := buildTargetTable("", "dc_cus"); got != "`dc_cus`" {
		t.Fatalf("buildTargetTable = %q", got)
	}
}
