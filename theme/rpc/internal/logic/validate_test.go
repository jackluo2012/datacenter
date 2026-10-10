package logic

import (
	"errors"
	"strings"
	"testing"
)

// TestValidateFieldsJsonPk 字段定义 pk 断言：主键必须恰有一个（0 个/2 个均拒绝）
func TestValidateFieldsJsonPk(t *testing.T) {
	//恰一个 pk：通过
	fields, err := ValidateFieldsJson(`[{"name":"id","type":"bigint","pk":true},{"name":"amt","type":"decimal"}]`)
	if err != nil {
		t.Fatalf("恰一个 pk 应通过: %v", err)
	}
	if len(fields) != 2 || fields[0].Name != "id" || !fields[0].Pk {
		t.Fatalf("解析结果不符: %+v", fields)
	}
	//0 个 pk：拒绝
	if _, err := ValidateFieldsJson(`[{"name":"id","type":"bigint"},{"name":"amt","type":"decimal"}]`); err == nil ||
		!errors.Is(err, ErrInvalidFieldsJson) || !strings.Contains(err.Error(), "恰有一个") {
		t.Fatalf("0 个 pk 应拒绝: %v", err)
	}
	//2 个 pk：拒绝
	if _, err := ValidateFieldsJson(`[{"name":"id","type":"bigint","pk":true},{"name":"uid","type":"bigint","pk":true}]`); err == nil ||
		!errors.Is(err, ErrInvalidFieldsJson) || !strings.Contains(err.Error(), "恰有一个") {
		t.Fatalf("2 个 pk 应拒绝: %v", err)
	}
	//空数组：拒绝
	if _, err := ValidateFieldsJson(`[]`); !errors.Is(err, ErrInvalidFieldsJson) {
		t.Fatalf("空字段定义应拒绝: %v", err)
	}
}

// TestValidateFieldsJsonInject 字段定义注入名拒绝与类型白名单
func TestValidateFieldsJsonInject(t *testing.T) {
	bad := []string{
		`[{"name":"a;drop table t","type":"string","pk":true}]`,              //分号
		`[{"name":"a b","type":"string","pk":true}]`,                         //空格
		`[{"name":"1abc","type":"string","pk":true}]`,                        //数字开头
		`[{"name":"a","type":"string"},{"name":"a","type":"int","pk":true}]`, //重复定义
		`[{"name":"a","type":"drop","pk":true}]`,                             //类型不在白名单
		`[{"name":"a","type":"INT","pk":true}]`,                              //类型大小写敏感
		`[{"name":"a","type":"string","pk":true},{"name":"b"}]`,              //缺 type
		`not json`,                                //非 JSON
		`{"name":"a","type":"string"}`,            //对象而非数组
		`[{"name":"","type":"string","pk":true}]`, //缺 name
	}
	for _, in := range bad {
		if _, err := ValidateFieldsJson(in); !errors.Is(err, ErrInvalidFieldsJson) {
			t.Fatalf("非法 fields_json 应拒绝: %q => %v", in, err)
		}
	}
}

// TestValidateDimJson 维度声明校验：空串放行；非空须为 [{dim_table,join_key}]
func TestValidateDimJson(t *testing.T) {
	if out, err := ValidateDimJson(""); err != nil || out != "" {
		t.Fatalf("空 dim_json 应原样放行: out=%q err=%v", out, err)
	}
	out, err := ValidateDimJson(`[{"dim_table":"dc_dim_shop","join_key":"shop_id"}]`)
	if err != nil || out == "" {
		t.Fatalf("合法 dim_json 应通过: out=%q err=%v", out, err)
	}
	for _, in := range []string{`not json`, `[{"dim_table":"t"}]`, `[{"join_key":"k"}]`, `{"dim_table":"t"}`} {
		if _, err := ValidateDimJson(in); !errors.Is(err, ErrInvalidDimJson) {
			t.Fatalf("非法 dim_json 应拒绝: %q => %v", in, err)
		}
	}
}

// TestValidateSelectSql 物化视图 select_sql 校验：SELECT 开头（大小写不敏感）且不含分号
func TestValidateSelectSql(t *testing.T) {
	for _, s := range []string{
		"select * from dwd_order_detail",
		"SELECT id, amt FROM dwd_order WHERE dt = '2026-01-01'",
		"  select 1  ",
		"SeLeCt 1",
	} {
		if err := ValidateSelectSql(s); err != nil {
			t.Fatalf("select_sql %q 应通过: %v", s, err)
		}
	}
	for _, s := range []string{
		"",                        //空
		"   ",                     //纯空白
		"delete from t",           //非 SELECT
		"update t set a=1",        //非 SELECT
		"insert into t values(1)", //非 SELECT
		"drop table t",            //非 SELECT
		"select 1; drop table t",  //多语句注入（含分号）
		"select 1;",               //尾分号同样拒绝
	} {
		if err := ValidateSelectSql(s); !errors.Is(err, ErrInvalidSelectSql) {
			t.Fatalf("select_sql %q 应拒绝: %v", s, err)
		}
	}
}

// TestValidateRefreshMode 刷新模式白名单
func TestValidateRefreshMode(t *testing.T) {
	if !ValidRefreshMode(RefreshModeManual) || !ValidRefreshMode(RefreshModeSchedule) {
		t.Fatal("manual/schedule 应在白名单")
	}
	for _, s := range []string{"", "auto", "MANUAL", "cron"} {
		if ValidRefreshMode(s) {
			t.Fatalf("refresh_mode %q 不应在白名单", s)
		}
	}
}
