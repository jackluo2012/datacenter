package logic

import (
	"testing"
)

// TestSqlTypeOf 类型映射表：string→varchar(255)/int→int/bigint→bigint/
// decimal→decimal(18,4)/bool→tinyint(1)/date→date/datetime→datetime(3)；
// 白名单外返回空串且 ValidFieldType 为假
func TestSqlTypeOf(t *testing.T) {
	cases := []struct {
		fieldType string
		want      string
	}{
		{FieldTypeString, "varchar(255)"},
		{FieldTypeInt, "int"},
		{FieldTypeBigint, "bigint"},
		{FieldTypeDecimal, "decimal(18,4)"},
		{FieldTypeBool, "tinyint(1)"},
		{FieldTypeDate, "date"},
		{FieldTypeDatetime, "datetime(3)"},
		{"float", ""},   //白名单外
		{"INT", ""},     //大小写敏感
		{"varchar", ""}, //白名单外
		{"", ""},        //空类型
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

// TestValidLayer 分层白名单：仅 ods/dwd/dws/ads/dim，大小写敏感
func TestValidLayer(t *testing.T) {
	for _, layer := range []string{"ods", "dwd", "dws", "ads", "dim"} {
		if !ValidLayer(layer) {
			t.Fatalf("layer %q 应在白名单", layer)
		}
	}
	for _, layer := range []string{"", "ODS", "dwd2", "dwt", "staging", "ods;drop", "o d s"} {
		if ValidLayer(layer) {
			t.Fatalf("layer %q 不应在白名单", layer)
		}
	}
}

// TestValidIdentifier 物理标识符白名单：小写字母开头，仅小写字母/数字/下划线
func TestValidIdentifier(t *testing.T) {
	for _, s := range []string{"a", "trade", "dwd_order_detail", "t_2024"} {
		if !ValidIdentifier(s) {
			t.Fatalf("%q 应为合法标识符", s)
		}
	}
	//注入名/非法名一律拒绝
	for _, s := range []string{"", "1abc", "_abc", "Abc", "a-b", "a b", "a;drop", "a`b", "order\"", "a'b"} {
		if ValidIdentifier(s) {
			t.Fatalf("%q 不应为合法标识符", s)
		}
	}
}

// TestValidFieldName 字段名白名单：字母开头，仅字母/数字/下划线（进 DDL 前把守）
func TestValidFieldName(t *testing.T) {
	for _, s := range []string{"id", "userName", "amount_2"} {
		if !ValidFieldName(s) {
			t.Fatalf("%q 应为合法字段名", s)
		}
	}
	for _, s := range []string{"", "1x", "a-b", "a b", "a;b", "a`b", "union select"} {
		if ValidFieldName(s) {
			t.Fatalf("%q 不应为合法字段名", s)
		}
	}
}
