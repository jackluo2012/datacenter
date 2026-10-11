package logic

import (
	"errors"
	"testing"
)

// 测试错误样本 SQL 生成：三类规则的语句形状与占位参数
func TestBuildErrorSampleSQL(t *testing.T) {
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
			name:     "not_null 样本",
			ruleType: RuleTypeNotNull,
			table:    "orders",
			column:   "user_id",
			wantSQL:  "select * from orders where `user_id` is null or `user_id` = '' limit 100",
		},
		{
			name:     "unique 重复组样本（带库前缀）",
			ruleType: RuleTypeUnique,
			db:       "mall",
			table:    "orders",
			column:   "order_no",
			wantSQL: "select t.* from mall.orders t join " +
				"(select `order_no` from mall.orders group by `order_no` having count(*) > 1) dup " +
				"on t.`order_no` = dup.`order_no` limit 100",
		},
		{
			name:       "value_range 越界样本",
			ruleType:   RuleTypeValueRange,
			table:      "orders",
			column:     "amount",
			paramsJson: `{"min":0,"max":10000}`,
			wantSQL:    "select * from orders where `amount` not between ? and ? limit 100",
			wantArgs:   []any{float64(0), float64(10000)},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotSQL, gotArgs, err := BuildErrorSampleSQL(c.ruleType, c.db, c.table, c.column, c.paramsJson)
			if err != nil {
				t.Fatalf("BuildErrorSampleSQL error: %v", err)
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

// 测试不支持样本抓取的规则类型与注入防护
func TestBuildErrorSampleSQLRejects(t *testing.T) {
	// 无违规行语义的类型
	for _, ruleType := range []string{RuleTypeRowCount, RuleTypeConsistency, RuleTypeFluctuation, "unknown"} {
		if _, _, err := BuildErrorSampleSQL(ruleType, "", "orders", "id", `{}`); !errors.Is(err, ErrNoErrorSample) {
			t.Fatalf("类型 %s 应返回 ErrNoErrorSample, got %v", ruleType, err)
		}
	}
	// 标识符注入一律拒绝
	if _, _, err := BuildErrorSampleSQL(RuleTypeNotNull, "", "orders; drop table x", "id", ""); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("表名注入应被拒绝, got %v", err)
	}
	if _, _, err := BuildErrorSampleSQL(RuleTypeNotNull, "", "orders", "id; delete from orders", ""); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("列名注入应被拒绝, got %v", err)
	}
	if _, _, err := BuildErrorSampleSQL(RuleTypeUnique, "mall; drop db", "orders", "id", ""); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("库名注入应被拒绝, got %v", err)
	}
	// value_range 缺参数提前报错
	if _, _, err := BuildErrorSampleSQL(RuleTypeValueRange, "", "orders", "amount", `{}`); !errors.Is(err, ErrInvalidRuleParams) {
		t.Fatalf("value_range 缺 min/max 应报参数错误, got %v", err)
	}
	// 空列名报错
	if _, _, err := BuildErrorSampleSQL(RuleTypeNotNull, "", "orders", "", ""); err == nil {
		t.Fatalf("空列名应报错")
	}
}
