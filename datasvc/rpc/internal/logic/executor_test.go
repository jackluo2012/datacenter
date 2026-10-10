package logic

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// 参数定义解析：正常 / 空 / 非法类型 / 重复参数 / 缺 name / 非数组
func TestParseParamDefs(t *testing.T) {
	// 空串视为无参数
	if defs, err := ParseParamDefs(""); err != nil || defs != nil {
		t.Fatalf("空 params_json 应返回 nil, got %v %v", defs, err)
	}
	defs, err := ParseParamDefs(`[{"name":"id","type":"number","required":true},{"name":"name","type":"string","default":"tom"}]`)
	if err != nil || len(defs) != 2 {
		t.Fatalf("正常定义解析失败: %v %v", defs, err)
	}
	// 非法类型
	if _, err := ParseParamDefs(`[{"name":"a","type":"int"}]`); err == nil {
		t.Fatal("非法类型应报错")
	}
	// 重复参数
	if _, err := ParseParamDefs(`[{"name":"a","type":"string"},{"name":"a","type":"string"}]`); err == nil {
		t.Fatal("重复参数应报错")
	}
	// 缺 name
	if _, err := ParseParamDefs(`[{"type":"string"}]`); err == nil {
		t.Fatal("缺 name 应报错")
	}
	// 非法 JSON
	if _, err := ParseParamDefs(`not-json`); err == nil {
		t.Fatal("非法 JSON 应报错")
	}
}

// 业务参数校验：缺 required / 多余参数 / default 补全 / 类型收敛
func TestValidateParams(t *testing.T) {
	defs, err := ParseParamDefs(`[
		{"name":"id","type":"number","required":true},
		{"name":"name","type":"string","required":false,"default":"tom"},
		{"name":"vip","type":"bool","required":false}
	]`)
	if err != nil {
		t.Fatalf("定义解析失败: %v", err)
	}
	// 缺 required 报错
	if _, err := ValidateParams(defs, `{}`); err == nil || !strings.Contains(err.Error(), "缺少必填参数") {
		t.Fatalf("缺 required 应报错, got: %v", err)
	}
	// 多余参数拒绝
	if _, err := ValidateParams(defs, `{"id":1,"hack":"1=1"}`); err == nil || !strings.Contains(err.Error(), "多余参数") {
		t.Fatalf("多余参数应拒绝, got: %v", err)
	}
	// 非必填缺省补 default
	vals, err := ValidateParams(defs, `{"id":1}`)
	if err != nil {
		t.Fatalf("合法参数不应报错: %v", err)
	}
	if vals["name"] != "tom" {
		t.Fatalf("缺省应补 default \"tom\", got: %v", vals["name"])
	}
	if _, ok := vals["vip"]; ok {
		t.Fatal("无 default 的非必填缺省时不应出现在结果中")
	}
	// 类型收敛：number→float64、bool→bool
	vals, err = ValidateParams(defs, `{"id":2,"vip":true,"name":"jerry"}`)
	if err != nil {
		t.Fatalf("合法参数不应报错: %v", err)
	}
	if n, ok := vals["id"].(float64); !ok || n != 2 {
		t.Fatalf("number 参数应为 float64 2, got: %T %v", vals["id"], vals["id"])
	}
	if vals["vip"] != true {
		t.Fatalf("bool 参数应为 true, got: %v", vals["vip"])
	}
	// 类型不符：number 传字符串
	if _, err := ValidateParams(defs, `{"id":"x"}`); err == nil || !strings.Contains(err.Error(), "类型错误") {
		t.Fatalf("number 传字符串应报类型错误, got: %v", err)
	}
	// 类型不符：bool 传数字
	if _, err := ValidateParams(defs, `{"id":1,"vip":1}`); err == nil {
		t.Fatalf("bool 传数字应报类型错误, got: %v", err)
	}
	// 业务参数不是对象
	if _, err := ValidateParams(defs, `[1,2]`); err == nil {
		t.Fatalf("数组业务参数应报错, got: %v", err)
	}
	// 业务参数非法 JSON
	if _, err := ValidateParams(defs, `{id:}`); err == nil {
		t.Fatalf("非法 JSON 业务参数应报错, got: %v", err)
	}
	// 空参数且定义里 required=false：只补 default
	vals, err = ValidateParams(nil, "")
	if err != nil || len(vals) != 0 {
		t.Fatalf("空定义+空参数应返回空 map, got: %v %v", vals, err)
	}
	// default 类型错误也要暴露
	badDefs, _ := ParseParamDefs(`[{"name":"n","type":"number","default":"oops"}]`)
	if _, err := ValidateParams(badDefs, ``); err == nil || !strings.Contains(err.Error(), "default 类型错误") {
		t.Fatalf("default 类型错误应报错, got: %v", err)
	}
}

// :name 占位符 → ? 替换与参数数组组装
func TestBuildSql(t *testing.T) {
	defs := []ParamDef{
		{Name: "id", Type: TypeNumber},
		{Name: "name", Type: TypeString},
	}
	values := map[string]interface{}{"id": 1.0, "name": "tom"}
	// 按出现顺序替换并组参数（含重复占位符）
	got, args, err := BuildSql(
		"select * from users where id = :id and name = :name and id > :id", defs, values)
	if err != nil {
		t.Fatalf("BuildSql 失败: %v", err)
	}
	want := "select * from users where id = ? and name = ? and id > ?"
	if got != want {
		t.Fatalf("替换结果不符:\nwant %q\ngot  %q", want, got)
	}
	wantArgs := []interface{}{1.0, "tom", 1.0}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("参数数组不符: want %v got %v", wantArgs, args)
	}
	// 白名单之外的占位符拒绝（防注入）
	if _, _, err := BuildSql("select * from t where a = :hack", defs, values); err == nil {
		t.Fatal("未声明的占位符应拒绝执行")
	}
	// 未传值的可选占位符以 NULL 占位
	optDefs := []ParamDef{{Name: "opt", Type: TypeString}}
	sql2, args2, err := BuildSql("select * from t where a = :opt", optDefs, map[string]interface{}{})
	if err != nil || sql2 != "select * from t where a = ?" || len(args2) != 1 || args2[0] != nil {
		t.Fatalf("可选占位符应替换为 ? 且参数为 nil: %q %v %v", sql2, args2, err)
	}
}

// SELECT 判定：大小写 / 前导空白与括号 / 写语句
func TestIsSelectStmt(t *testing.T) {
	cases := []struct {
		sql  string
		want bool
	}{
		{"select 1", true},
		{"  SELECT * FROM t", true},
		{"(select id from t limit 1)", true},
		{"insert into t values (1)", false},
		{"UPDATE t set a=1", false},
		{"delete from t", false},
		{"DROP table t", false},
		{"", false},
		{"   ", false},
	}
	for _, c := range cases {
		if got := IsSelectStmt(c.sql); got != c.want {
			t.Fatalf("IsSelectStmt(%q) = %v, want %v", c.sql, got, c.want)
		}
	}
}

// field_mask 脱敏：命中列名（忽略大小写）值替换为 ****，其余不变
func TestMaskRows(t *testing.T) {
	rows := []map[string]interface{}{
		{"id": 1, "phone": "13800000001", "name": "a"},
		{"id": 2, "phone": "13800000002", "name": "b"},
	}
	// 命中脱敏
	MaskRows(rows, "phone")
	if rows[0]["phone"] != maskValue || rows[1]["phone"] != maskValue {
		t.Fatalf("phone 列应被脱敏: %v", rows)
	}
	if rows[0]["id"] != 1 || rows[0]["name"] != "a" {
		t.Fatal("未命中列不应被改动")
	}
	// 大小写不敏感命中（结果列 Phone vs mask phone）
	rows2 := []map[string]interface{}{{"Phone": "13800000001", "ID": 3}}
	MaskRows(rows2, "phone,id")
	if rows2[0]["Phone"] != maskValue || rows2[0]["ID"] != maskValue {
		t.Fatalf("大小写命中失败: %v", rows2)
	}
	// 空 mask 不改动
	rows3 := []map[string]interface{}{{"phone": "13800000001"}}
	MaskRows(rows3, "  ")
	if rows3[0]["phone"] != "13800000001" {
		t.Fatal("空 mask 不应改动结果")
	}
	// 空 rows 安全
	MaskRows(nil, "phone")
}

// 缓存 key：同参数同 key（与传入顺序无关）、不同 api/参数不同 key
func TestCacheKey(t *testing.T) {
	v1 := map[string]interface{}{"id": 1.0, "name": "tom"}
	v2 := map[string]interface{}{"name": "tom", "id": 1.0} //仅顺序不同
	if CacheKey(7, v1) != CacheKey(7, v2) {
		t.Fatal("相同参数（不同顺序）应生成相同 key")
	}
	if CacheKey(7, v1) == CacheKey(8, v1) {
		t.Fatal("不同 api_id 应生成不同 key")
	}
	if CacheKey(7, v1) == CacheKey(7, map[string]interface{}{"id": 2.0, "name": "tom"}) {
		t.Fatal("不同参数应生成不同 key")
	}
	if !strings.HasPrefix(CacheKey(7, v1), "dsvc#api#7#") {
		t.Fatalf("key 应符合 dsvc#api#{api_id}# 前缀: %q", CacheKey(7, v1))
	}
}

// IP 白名单规范化：去空白、丢空项
func TestNormalizeWhitelist(t *testing.T) {
	if got := normalizeWhitelist(" 10.0.0.1 , , 192.168.0.1 "); got != "10.0.0.1,192.168.0.1" {
		t.Fatalf("规范化结果不符: %q", got)
	}
	if got := normalizeWhitelist(""); got != "" {
		t.Fatalf("空白名单应为空串: %q", got)
	}
}

// 缓存结果集反序列化
func TestDecodeRows(t *testing.T) {
	rows, err := decodeRows(`[{"a":1},{"a":"x"}]`)
	if err != nil || len(rows) != 2 {
		t.Fatalf("decodeRows 失败: %v %v", rows, err)
	}
	if _, err := decodeRows(`not-json`); err == nil {
		t.Fatal("非法 JSON 应报错")
	}
	if _, err := decodeRows(`{"a":1}`); err != nil {
		// 对象形式非数组：json 可解到 []map 会失败
		t.Log("对象解到数组返回错误属预期")
	}
}

// 归纳校验：ExecuteApi 业务错误可用 %w 包装后保持 errors.Is 语义
func TestSentinelErrors(t *testing.T) {
	err := fmt.Errorf("%w: 授权已过期", ErrNotGranted)
	if !errors.Is(err, ErrNotGranted) {
		t.Fatal("包装后的哨兵错误应保持 errors.Is 语义")
	}
	err = fmt.Errorf("%w: ts 与服务器时间偏差超过 300 秒", ErrSignInvalid)
	if !errors.Is(err, ErrSignInvalid) {
		t.Fatal("包装后的 ErrSignInvalid 应保持 errors.Is 语义")
	}
}

// 目标库返回值归一化：[]byte → string（其余保持）
func TestNormalizeDbValue(t *testing.T) {
	if got := normalizeDbValue([]byte("abc")); got != "abc" {
		t.Fatalf("[]byte 应转 string, got %T %v", got, got)
	}
	if got := normalizeDbValue(1.5); got != 1.5 {
		t.Fatalf("float64 应保持原样, got %v", got)
	}
}
