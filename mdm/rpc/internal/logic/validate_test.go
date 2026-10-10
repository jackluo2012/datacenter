package logic

import (
	"encoding/json"
	"strings"
	"testing"
)

// mustJSON map 转紧凑 JSON（断言规范化输出用）
func mustJSON(t *testing.T, v map[string]any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %v: %v", v, err)
	}
	return string(b)
}

// TestValidateFieldsJson 模型字段定义校验：合法（含附加键）/缺 name/空 name/name 非法/
// 内置列冲突/name 重复/缺 type/type 白名单外/非数组/元素非对象
func TestValidateFieldsJson(t *testing.T) {
	cases := []struct {
		name       string
		fieldsJson string
		wantErr    bool
	}{
		{"合法定义", `[{"name":"name","type":"string","required":true},{"name":"age","type":"int"},{"name":"balance","type":"decimal"},{"name":"vip","type":"bool"},{"name":"birthday","type":"date"}]`, false},
		{"附加键 length/comment 合法保留", `[{"name":"name","type":"string","length":64,"required":true,"comment":"姓名"}]`, false},
		{"空数组允许", `[]`, false},
		{"缺 name", `[{"type":"string"}]`, true},
		{"空 name", `[{"name":"  ","type":"string"}]`, true},
		{"name 非法字符（注入）", `[{"name":"name` + "`" + ` drop table x","type":"string"}]`, true},
		{"name 数字开头", `[{"name":"1name","type":"string"}]`, true},
		{"name 与内置列 biz_code 冲突", `[{"name":"biz_code","type":"string"}]`, true},
		{"name 与内置列 created_at 冲突", `[{"name":"CREATED_AT","type":"string"}]`, true},
		{"name 重复", `[{"name":"age","type":"int"},{"name":"age","type":"string"}]`, true},
		{"缺 type", `[{"name":"age"}]`, true},
		{"type 白名单外", `[{"name":"age","type":"float"}]`, true},
		{"type 大小写敏感", `[{"name":"age","type":"INT"}]`, true},
		{"顶层不是数组", `{"name":"age","type":"int"}`, true},
		{"元素不是对象", `["age"]`, true},
		{"非法 JSON", `[{"name":}]`, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fields, err := ValidateFieldsJson(c.fieldsJson)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidateFieldsJson(%s) err = %v, wantErr = %v", c.fieldsJson, err, c.wantErr)
			}
			if !c.wantErr && fields == nil {
				t.Fatalf("合法定义应返回字段列表")
			}
		})
	}

	//附加键保留断言（schema 约定 length/required/comment 为合法可选键）
	fields, err := ValidateFieldsJson(`[{"name":"name","type":"string","length":64,"required":true,"comment":"姓名"}]`)
	if err != nil {
		t.Fatalf("附加键定义应合法: %v", err)
	}
	if fields[0].Length != 64 || !fields[0].Required || fields[0].Comment != "姓名" {
		t.Fatalf("附加键丢失: %+v", fields[0])
	}
}

// TestValidateDataJson 记录数据字段级校验：required 缺失/null/空串、类型收敛
// （string 原样、int/decimal/bool 字符串收敛、date 格式）、多余字段拒绝、规范化输出
func TestValidateDataJson(t *testing.T) {
	fields, err := ValidateFieldsJson(`[
		{"name":"name","type":"string","required":true},
		{"name":"age","type":"int","required":true},
		{"name":"balance","type":"decimal"},
		{"name":"vip","type":"bool"},
		{"name":"birthday","type":"date"}
	]`)
	if err != nil {
		t.Fatalf("模型定义非法: %v", err)
	}

	cases := []struct {
		name       string
		dataJson   string
		wantErr    bool
		wantOutput string //wantErr=false 时断言规范化结果（空串不断言）
	}{
		//合法 + 类型收敛
		{"全字符串数值", `{"name":"张三","age":"42","balance":"3.14","vip":"true","birthday":"2026-10-10"}`,
			false, mustJSON(t, map[string]any{"name": "张三", "age": 42, "balance": 3.14, "vip": true, "birthday": "2026-10-10"})},
		{"全 JSON 数值", `{"name":"张三","age":42,"balance":3.14,"vip":true,"birthday":"2026-10-10"}`,
			false, mustJSON(t, map[string]any{"name": "张三", "age": 42, "balance": 3.14, "vip": true, "birthday": "2026-10-10"})},
		{"int 整型浮点收敛", `{"name":"a","age":42.0}`, false, mustJSON(t, map[string]any{"name": "a", "age": 42})},
		{"bool 数字字符串收敛", `{"name":"a","age":1,"vip":"1"}`, false, mustJSON(t, map[string]any{"name": "a", "age": 1, "vip": true})},
		{"非必填缺失允许", `{"name":"a","age":1}`, false, mustJSON(t, map[string]any{"name": "a", "age": 1})},
		{"非必填 null 视为缺失", `{"name":"a","age":1,"balance":null}`, false, mustJSON(t, map[string]any{"name": "a", "age": 1})},
		//required 校验
		{"必填缺失", `{"age":1}`, true, ""},
		{"必填 null", `{"name":null,"age":1}`, true, ""},
		{"必填空串", `{"name":"","age":1}`, true, ""},
		{"必填全缺", `{}`, true, ""},
		//类型收敛失败
		{"int 非数字字符串", `{"name":"a","age":"abc"}`, true, ""},
		{"int 小数拒绝", `{"name":"a","age":4.5}`, true, ""},
		{"int 布尔拒绝", `{"name":"a","age":true}`, true, ""},
		{"decimal 非数值", `{"name":"a","age":1,"balance":"x3.14"}`, true, ""},
		{"bool 非法字符串", `{"name":"a","age":1,"vip":"yes"}`, true, ""},
		{"bool 数字拒绝", `{"name":"a","age":1,"vip":1}`, true, ""},
		{"date 非字符串", `{"name":"a","age":1,"birthday":20261010}`, true, ""},
		{"date 带时间拒绝", `{"name":"a","age":1,"birthday":"2026-10-10 00:00:00"}`, true, ""},
		{"date 非零填充拒绝", `{"name":"a","age":1,"birthday":"2026-1-2"}`, true, ""},
		{"date 非法日期拒绝", `{"name":"a","age":1,"birthday":"2026-13-40"}`, true, ""},
		{"string 非字符串拒绝", `{"name":123,"age":1}`, true, ""},
		//多余字段拒绝
		{"多余字段拒绝", `{"name":"a","age":1,"hack":"x"}`, true, ""},
		{"纯多余字段拒绝", `{"hack":"x"}`, true, ""},
		//结构
		{"非对象拒绝", `["a"]`, true, ""},
		{"非法 JSON 拒绝", `{"name":}`, true, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ValidateDataJson(fields, c.dataJson)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidateDataJson(%s) err = %v, wantErr = %v", c.dataJson, err, c.wantErr)
			}
			if !c.wantErr && c.wantOutput != "" && got != c.wantOutput {
				t.Fatalf("规范化输出不符:\n got = %s\nwant = %s", got, c.wantOutput)
			}
		})
	}
}

// TestNormalizeTableName 表名规范化：为空自动生成 dc_mdm_data_<model_code>、
// 显式表名保留、非法表名/模型编码拒绝、超长拒绝
func TestNormalizeTableName(t *testing.T) {
	cases := []struct {
		name      string
		modelCode string
		tableName string
		want      string
		wantErr   bool
	}{
		{"为空自动生成", "customer", "", "dc_mdm_data_customer", false},
		{"显式表名保留", "customer", "dc_crm_customer", "dc_crm_customer", false},
		{"表名前后空白收敛", "customer", "  dc_crm_customer  ", "dc_crm_customer", false},
		{"非法表名（注入）", "customer", "x; drop table t", "", true},
		{"非法表名（大写开头）", "customer", "Customer", "", true},
		{"非法表名（连字符）", "customer", "dc-crm", "", true},
		{"非法表名（数字开头）", "customer", "1t", "", true},
		{"模型编码非法且表名为空", "Customer", "", "", true},
		{"模型编码非法含点", "cus.tom", "", "", true},
		{"表名超 128 字符", "customer", strings.Repeat("a", 129), "", true},
		{"表名恰 128 字符", "customer", strings.Repeat("a", 128), strings.Repeat("a", 128), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NormalizeTableName(c.modelCode, c.tableName)
			if (err != nil) != c.wantErr {
				t.Fatalf("NormalizeTableName(%q,%q) err = %v, wantErr = %v", c.modelCode, c.tableName, err, c.wantErr)
			}
			if !c.wantErr && got != c.want {
				t.Fatalf("NormalizeTableName(%q,%q) = %q, want %q", c.modelCode, c.tableName, got, c.want)
			}
		})
	}
}
