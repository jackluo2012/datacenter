package logic

import (
	"errors"
	"strings"
	"testing"
)

// ddlFields 建表 DDL 测试用字段组：bigint 主键 + 各类型字段
var ddlFields = []FieldDef{
	{Name: "id", Type: FieldTypeBigint, Pk: true},
	{Name: "user_name", Type: FieldTypeString},
	{Name: "amount", Type: FieldTypeDecimal},
	{Name: "is_vip", Type: FieldTypeBool},
	{Name: "biz_date", Type: FieldTypeDate},
	{Name: "created", Type: FieldTypeDatetime},
	{Name: "cnt", Type: FieldTypeInt},
}

// TestBuildPhysicalTable 物理表名 = <layer>_<table>
func TestBuildPhysicalTable(t *testing.T) {
	if got := buildPhysicalTable("dwd", "order_detail"); got != "dwd_order_detail" {
		t.Fatalf("buildPhysicalTable = %q, want dwd_order_detail", got)
	}
}

// TestBuildQualifiedTable 限定名：db 非空时 `db`.`table`，否则 `table`
func TestBuildQualifiedTable(t *testing.T) {
	if got := buildQualifiedTable("", "t1"); got != "`t1`" {
		t.Fatalf("buildQualifiedTable(\"\", t1) = %q", got)
	}
	if got := buildQualifiedTable("dw", "t1"); got != "`dw`.`t1`" {
		t.Fatalf("buildQualifiedTable(dw, t1) = %q", got)
	}
}

// TestBuildCreateTableDDL fields→DDL 生成：幂等前缀/物理表名/类型映射/pk 子句逐项断言
func TestBuildCreateTableDDL(t *testing.T) {
	ddl, err := buildCreateTableDDL("`dw`.`dwd_order_detail`", ddlFields)
	if err != nil {
		t.Fatalf("生成 DDL 失败: %v", err)
	}
	for _, want := range []string{
		"create table if not exists `dw`.`dwd_order_detail`", //幂等前缀 + 物理表名
		"`id` bigint NOT NULL",                               //主键列 NOT NULL
		"`user_name` varchar(255) DEFAULT NULL",
		"`amount` decimal(18,4) DEFAULT NULL",
		"`is_vip` tinyint(1) DEFAULT NULL",
		"`biz_date` date DEFAULT NULL",
		"`created` datetime(3) DEFAULT NULL",
		"`cnt` int DEFAULT NULL",
		"primary key(`id`)", //主键恰为 pk 字段
	} {
		if !strings.Contains(ddl, want) {
			t.Fatalf("DDL 缺少片段 %q:\n%s", want, ddl)
		}
	}
	//主键子句仅出现一次
	if strings.Count(ddl, "primary key(") != 1 {
		t.Fatalf("primary key 子句应恰有一个:\n%s", ddl)
	}
}

// TestBuildCreateTableDDLPkAssert pk 断言：0 个主键拒绝建表
func TestBuildCreateTableDDLPkAssert(t *testing.T) {
	noPk := []FieldDef{{Name: "a", Type: FieldTypeString}, {Name: "b", Type: FieldTypeInt}}
	if _, err := buildCreateTableDDL("`t`", noPk); !errors.Is(err, ErrInvalidFieldsJson) {
		t.Fatalf("无主键应拒绝: %v", err)
	}
	empty := []FieldDef{}
	if _, err := buildCreateTableDDL("`t`", empty); !errors.Is(err, ErrInvalidFieldsJson) {
		t.Fatalf("空字段应拒绝: %v", err)
	}
}

// TestBuildCreateTableDDLInjectName 注入名拒绝：列名/限定名未过白名单一律拒绝
func TestBuildCreateTableDDLInjectName(t *testing.T) {
	injectFields := []FieldDef{
		{Name: "id` int, hack varchar(10) --", Type: FieldTypeBigint, Pk: true},
	}
	if _, err := buildCreateTableDDL("`t`", injectFields); !errors.Is(err, ErrInvalidFieldsJson) {
		t.Fatalf("注入列名应拒绝: %v", err)
	}
	//限定名带引号/分号：白名单（^[A-Za-z0-9_.]+$ 剥反引号后）拒绝
	if _, err := buildCreateTableDDL("`t`; drop table x", ddlFields); err == nil {
		t.Fatal("注入限定名应拒绝")
	}
	if _, err := buildCreateTableDDL("t`)", ddlFields); err == nil {
		t.Fatal("非法限定名应拒绝")
	}
	//合法限定名（含 db 限定）放行
	if _, err := buildCreateTableDDL("`dw`.`ods_trade`", ddlFields); err != nil {
		t.Fatalf("合法限定名应放行: %v", err)
	}
}

// TestBuildCreateViewDDL 物化视图刷新 DDL：create or replace view ... as select...
func TestBuildCreateViewDDL(t *testing.T) {
	ddl := buildCreateViewDDL("`dw`.`mv_order_daily`", "  select dt, sum(amount) from dwd_order group by dt ")
	want := "create or replace view `dw`.`mv_order_daily` as select dt, sum(amount) from dwd_order group by dt"
	if ddl != want {
		t.Fatalf("buildCreateViewDDL = %q, want %q", ddl, want)
	}
}

// TestSummarizeSql 摘要截断
func TestSummarizeSql(t *testing.T) {
	long := strings.Repeat("a", 400)
	got := summarizeSql(long, 100)
	if len(got) != 103 || !strings.HasSuffix(got, "...") {
		t.Fatalf("超长摘要应截断加省略号，len=%d", len(got))
	}
	if summarizeSql("select 1", 100) != "select 1" {
		t.Fatal("短 SQL 应原样返回")
	}
}

// TestValidateFieldsJsonToDDL 校验→生成端到端：合法定义可出 DDL，注入定义在校验层拦截
func TestValidateFieldsJsonToDDL(t *testing.T) {
	fields, err := ValidateFieldsJson(`[{"name":"id","type":"bigint","pk":true,"comment":"主键"},{"name":"amt","type":"decimal","dim":false}]`)
	if err != nil {
		t.Fatalf("合法定义应通过: %v", err)
	}
	ddl, err := buildCreateTableDDL(buildQualifiedTable("", buildPhysicalTable(LayerAds, "user_portrait")), fields)
	if err != nil {
		t.Fatalf("生成 DDL 失败: %v", err)
	}
	if !strings.Contains(ddl, "create table if not exists `ads_user_portrait`") ||
		!strings.Contains(ddl, "primary key(`id`)") {
		t.Fatalf("端到端 DDL 不符:\n%s", ddl)
	}
	//注入名在校验层即拒绝，走不到 DDL
	if _, err := ValidateFieldsJson(`[{"name":"id","type":"bigint","pk":true},{"name":"x); drop table t--","type":"string"}]`); !errors.Is(err, ErrInvalidFieldsJson) {
		t.Fatalf("注入字段名应在校验层拒绝: %v", err)
	}
}
