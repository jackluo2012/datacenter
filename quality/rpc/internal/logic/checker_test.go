package logic

import (
	"errors"
	"reflect"
	"testing"
)

// 测试 6 种规则的检查 SQL 生成（含库前缀与占位参数）
func TestBuildCheckSQL(t *testing.T) {
	cases := []struct {
		name       string
		ruleType   string
		db         string
		table      string
		column     string
		paramsJson string
		wantSQL    string
		wantArgs   []any
	}{
		{
			name:     "row_count",
			ruleType: RuleTypeRowCount,
			table:    "orders",
			wantSQL:  "select count(*) from orders",
		},
		{
			name:     "not_null",
			ruleType: RuleTypeNotNull,
			table:    "orders",
			column:   "user_id",
			wantSQL:  "select count(*) from orders where `user_id` is null or `user_id` = ''",
		},
		{
			name:     "unique",
			ruleType: RuleTypeUnique,
			table:    "orders",
			column:   "order_no",
			wantSQL:  "select count(*) from (select `order_no` from orders group by `order_no` having count(*) > 1) t",
		},
		{
			name:       "value_range 带库前缀与占位参数",
			ruleType:   RuleTypeValueRange,
			db:         "mall",
			table:      "orders",
			column:     "amount",
			paramsJson: `{"min":0,"max":10000}`,
			wantSQL:    "select count(*) from mall.orders where `amount` not between ? and ?",
			wantArgs:   []any{float64(0), float64(10000)},
		},
		{
			name:       "consistency",
			ruleType:   RuleTypeConsistency,
			db:         "mall",
			table:      "orders",
			column:     "user_id",
			paramsJson: `{"ref_table":"users","ref_column":"id"}`,
			wantSQL:    "select count(*) from mall.orders t left join mall.users r on t.`user_id` = r.`id` where r.`id` is null",
		},
		{
			name:       "fluctuation",
			ruleType:   RuleTypeFluctuation,
			table:      "orders",
			paramsJson: `{"baseline":1000,"ratio_pct":5}`,
			wantSQL:    "select count(*) from orders",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotSQL, gotArgs, err := BuildCheckSQL(c.ruleType, c.db, c.table, c.column, c.paramsJson)
			if err != nil {
				t.Fatalf("BuildCheckSQL error: %v", err)
			}
			if gotSQL != c.wantSQL {
				t.Fatalf("sql = %q, want %q", gotSQL, c.wantSQL)
			}
			if len(gotArgs) != len(c.wantArgs) {
				t.Fatalf("args = %v, want %v", gotArgs, c.wantArgs)
			}
			for i := range gotArgs {
				if gotArgs[i] != c.wantArgs[i] {
					t.Fatalf("args[%d] = %v, want %v", i, gotArgs[i], c.wantArgs[i])
				}
			}
		})
	}
}

// 测试 SQL 注入防护：表名/列名/库名未通过白名单的必须被拒
func TestBuildCheckSQLRejectsInjection(t *testing.T) {
	badTargets := []string{
		"users; drop table x",
		"users;drop table x",
		"users--",
		"users or 1=1",
		"`users`",
		"users where 1=1",
		"user name", // 空格
		"",          // 空表名
	}
	for _, target := range badTargets {
		if _, _, err := BuildCheckSQL(RuleTypeRowCount, "", target, "", ""); !errors.Is(err, ErrInvalidIdentifier) {
			t.Fatalf("表名 %q 应被拒绝, got err = %v", target, err)
		}
	}
	// 列名注入
	if _, _, err := BuildCheckSQL(RuleTypeNotNull, "", "orders", "amount; drop table x", ""); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("列名注入应被拒绝, got err = %v", err)
	}
	// 库名注入（含点号与分号均不允许）
	if _, _, err := BuildCheckSQL(RuleTypeRowCount, "mall; drop table x", "orders", "", ""); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("库名注入应被拒绝, got err = %v", err)
	}
	if _, _, err := BuildCheckSQL(RuleTypeRowCount, "mall.orders", "orders", "", ""); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("库名含点号应被拒绝（前缀由 db 参数承担）, got err = %v", err)
	}
	// consistency 的 ref_table/ref_column 注入
	if _, _, err := BuildCheckSQL(RuleTypeConsistency, "", "orders", "user_id", `{"ref_table":"users; drop table x","ref_column":"id"}`); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("ref_table 注入应被拒绝, got err = %v", err)
	}
	if _, _, err := BuildCheckSQL(RuleTypeConsistency, "", "orders", "user_id", `{"ref_table":"users","ref_column":"id; delete from users"}`); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("ref_column 注入应被拒绝, got err = %v", err)
	}
}

// 测试 value_range/consistency/fluctuation 缺参数时提前报错
func TestBuildCheckSQLRequiresParams(t *testing.T) {
	cases := []struct {
		name       string
		ruleType   string
		column     string
		paramsJson string
	}{
		{"value_range 缺 min/max", RuleTypeValueRange, "amount", `{}`},
		{"value_range 参数非数字", RuleTypeValueRange, "amount", `{"min":"a","max":10}`},
		{"consistency 缺 ref", RuleTypeConsistency, "user_id", `{"ref_table":"users"}`},
		{"fluctuation 缺 baseline", RuleTypeFluctuation, "", `{"ratio_pct":5}`},
		{"not_null 缺列", RuleTypeNotNull, "", ""},
	}
	for _, c := range cases {
		if _, _, err := BuildCheckSQL(c.ruleType, "", "orders", c.column, c.paramsJson); err == nil {
			t.Fatalf("%s 应报错", c.name)
		}
	}
}

// 测试判定口径：命中数/行数的通过规则
func TestJudgeCheck(t *testing.T) {
	// not_null/unique/value_range/consistency：0 通过、非 0 不通过
	for _, ruleType := range []string{RuleTypeNotNull, RuleTypeUnique, RuleTypeValueRange, RuleTypeConsistency} {
		pass, _, err := JudgeCheck(ruleType, 0, `{"min":0,"max":1}`, "")
		if err != nil || !pass {
			t.Fatalf("%s 实测 0 应通过, pass=%v err=%v", ruleType, pass, err)
		}
		pass, detail, err := JudgeCheck(ruleType, 3, `{"min":0,"max":1}`, "")
		if err != nil || pass {
			t.Fatalf("%s 实测 3 应不通过, pass=%v err=%v", ruleType, pass, err)
		}
		if detail == "" {
			t.Fatalf("%s 应返回 detail_json", ruleType)
		}
	}
}

// 测试 row_count 阈值判定：缺省要求表非空，提供 threshold_json 时按 op 比较
func TestJudgeRowCountThreshold(t *testing.T) {
	cases := []struct {
		name          string
		count         int64
		thresholdJson string
		wantPass      bool
	}{
		{"缺省阈值-非空表通过", 100, "", true},
		{"缺省阈值-空表不通过", 0, "", false},
		{"gt 满足", 100, `{"op":"gt","value":99}`, true},
		{"gt 不满足", 100, `{"op":"gt","value":100}`, false},
		{"gte 满足", 100, `{"op":"gte","value":100}`, true},
		{"lt 满足", 5, `{"op":"lt","value":10}`, true},
		{"eq 不满足", 5, `{"op":"eq","value":10}`, false},
		{"between 满足", 50, `{"op":"between","value":[10,100]}`, true},
		{"between 逆序自动纠正", 50, `{"op":"between","value":[100,10]}`, true},
		{"between 越界", 101, `{"op":"between","value":[10,100]}`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pass, _, err := JudgeCheck(RuleTypeRowCount, c.count, "", c.thresholdJson)
			if err != nil {
				t.Fatalf("JudgeCheck error: %v", err)
			}
			if pass != c.wantPass {
				t.Fatalf("count=%d pass = %v, want %v", c.count, pass, c.wantPass)
			}
		})
	}
	// 非法阈值 op 一律不通过（不 panic）
	if pass, _, _ := JudgeCheck(RuleTypeRowCount, 100, "", `{"op":"hack","value":1}`); pass {
		t.Fatalf("未知 op 应不通过")
	}
}

// 测试波动率判定：偏差比例与阈值比较、基线为 0 的约定
func TestJudgeFluctuation(t *testing.T) {
	cases := []struct {
		name       string
		current    int64
		paramsJson string
		wantPass   bool
	}{
		{"偏差 2% 在 5% 内", 1020, `{"baseline":1000,"ratio_pct":5}`, true},
		{"偏差恰好等于阈值", 1050, `{"baseline":1000,"ratio_pct":5}`, true},
		{"偏差超阈值", 1060, `{"baseline":1000,"ratio_pct":5}`, false},
		{"行数下降同样超限", 900, `{"baseline":1000,"ratio_pct":5}`, false},
		{"行数下降在限内", 980, `{"baseline":1000,"ratio_pct":5}`, true},
		{"基线为 0 当前行也为 0", 0, `{"baseline":0,"ratio_pct":5}`, true},
		{"基线为 0 当前行非 0", 10, `{"baseline":0,"ratio_pct":5}`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pass, detail, err := JudgeCheck(RuleTypeFluctuation, c.current, c.paramsJson, "")
			if err != nil {
				t.Fatalf("JudgeCheck error: %v", err)
			}
			if pass != c.wantPass {
				t.Fatalf("current=%d pass = %v, want %v", c.current, pass, c.wantPass)
			}
			if detail == "" {
				t.Fatalf("应返回 detail_json")
			}
		})
	}
	// 缺参数报错
	if _, _, err := JudgeCheck(RuleTypeFluctuation, 10, `{}`, ""); !errors.Is(err, ErrInvalidRuleParams) {
		t.Fatalf("缺 baseline/ratio_pct 应报参数错误, got %v", err)
	}
}

// 测试创建规则的静态校验（类型白名单 + 按类型必填性）
func TestValidateRule(t *testing.T) {
	valid := []struct {
		ruleType   string
		column     string
		paramsJson string
	}{
		{RuleTypeRowCount, "", ""},
		{RuleTypeNotNull, "user_id", ""},
		{RuleTypeUnique, "order_no", ""},
		{RuleTypeValueRange, "amount", `{"min":0,"max":100}`},
		{RuleTypeConsistency, "user_id", `{"ref_table":"users","ref_column":"id"}`},
		{RuleTypeFluctuation, "", `{"baseline":100,"ratio_pct":10}`},
	}
	for _, c := range valid {
		if err := ValidateRule(c.ruleType, "orders", c.column, c.paramsJson); err != nil {
			t.Fatalf("%s 合法用例不应报错, got %v", c.ruleType, err)
		}
	}

	invalid := []struct {
		name        string
		ruleType    string
		targetTable string
		column      string
		paramsJson  string
		wantErr     error
	}{
		{"类型不在白名单", "select", "orders", "", "", ErrInvalidRuleType},
		{"缺目标表", RuleTypeNotNull, "", "id", "", ErrTargetTableRequired},
		{"目标表注入", RuleTypeRowCount, "users; drop table x", "", "", ErrInvalidIdentifier},
		{"not_null 缺列", RuleTypeNotNull, "orders", "", "", ErrInvalidRuleParams},
		{"unique 缺列", RuleTypeUnique, "orders", "", "", ErrInvalidRuleParams},
		{"value_range 缺列", RuleTypeValueRange, "orders", "", `{"min":0,"max":1}`, ErrInvalidRuleParams},
		{"value_range 缺参数", RuleTypeValueRange, "orders", "amount", `{}`, ErrInvalidRuleParams},
		{"value_range min>max", RuleTypeValueRange, "orders", "amount", `{"min":10,"max":1}`, ErrInvalidRuleParams},
		{"consistency 缺 ref_column", RuleTypeConsistency, "orders", "user_id", `{"ref_table":"users"}`, ErrInvalidRuleParams},
		{"fluctuation 缺 ratio_pct", RuleTypeFluctuation, "orders", "", `{"baseline":100}`, ErrInvalidRuleParams},
		{"fluctuation 负阈值", RuleTypeFluctuation, "orders", "", `{"baseline":100,"ratio_pct":-1}`, ErrInvalidRuleParams},
	}
	for _, c := range invalid {
		err := ValidateRule(c.ruleType, c.targetTable, c.column, c.paramsJson)
		if err == nil {
			t.Fatalf("%s 应报错", c.name)
		}
		if !errors.Is(err, c.wantErr) {
			t.Fatalf("%s 错误不匹配: got %v, want %v", c.name, err, c.wantErr)
		}
	}
}

// 测试规则 id 串解析：容错空白、去重保序、拒非法片段
func TestParseRuleIds(t *testing.T) {
	got, err := ParseRuleIds(" 3, 1 ,2,1 ")
	if err != nil {
		t.Fatalf("ParseRuleIds error: %v", err)
	}
	if !reflect.DeepEqual(got, []int64{3, 1, 2}) {
		t.Fatalf("ids = %v, want [3 1 2]", got)
	}
	if _, err := ParseRuleIds(""); !errors.Is(err, ErrEmptyRuleIds) {
		t.Fatalf("空串应报 ErrEmptyRuleIds, got %v", err)
	}
	if _, err := ParseRuleIds(",,"); !errors.Is(err, ErrEmptyRuleIds) {
		t.Fatalf("纯分隔符应报 ErrEmptyRuleIds, got %v", err)
	}
	if _, err := ParseRuleIds("1,a,3"); !errors.Is(err, ErrInvalidRuleIds) {
		t.Fatalf("非法片段应报 ErrInvalidRuleIds, got %v", err)
	}
	if _, err := ParseRuleIds("1,0,-2"); !errors.Is(err, ErrInvalidRuleIds) {
		t.Fatalf("非正整数应报 ErrInvalidRuleIds, got %v", err)
	}
}

// 测试标识符白名单正则的边界
func TestValidIdentifier(t *testing.T) {
	valid := []string{"users", "t_1", "_tmp", "OrderItems", "a"}
	for _, s := range valid {
		if !ValidIdentifier(s) {
			t.Fatalf("%q 应为合法标识符", s)
		}
	}
	invalid := []string{"", "users; drop table x", "a b", "a-b", "a.b", "`a`", "'a'", "a*", "中文名"}
	for _, s := range invalid {
		if ValidIdentifier(s) {
			t.Fatalf("%q 应为非法标识符", s)
		}
	}
}

// 测试库前缀生成
func TestQualifyTable(t *testing.T) {
	got, err := QualifyTable("mall", "orders")
	if err != nil || got != "mall.orders" {
		t.Fatalf("QualifyTable = %q, %v", got, err)
	}
	got, err = QualifyTable("", "orders")
	if err != nil || got != "orders" {
		t.Fatalf("空库名应返回裸表名, got %q, %v", got, err)
	}
	if _, err := QualifyTable("mall", "orders; drop table x"); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("表名注入应被拒绝, got %v", err)
	}
}
