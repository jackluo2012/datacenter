package logic

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"datacenter/quality/model"
)

// 测试健康分纯函数：标准公式、四舍五入、0/100 边界与夹取
func TestHealthScore(t *testing.T) {
	cases := []struct {
		name                 string
		nullRate, dup, extra float64
		want                 int32
	}{
		{"零瑕疵满分", 0, 0, 0, 100},
		{"全空扣40", 1, 0, 0, 60},
		{"主键全重复扣30", 0, 1, 0, 70},
		{"空值率一半", 0.5, 0, 0, 80},
		{"组合扣分", 0.25, 0.1, 0, 87}, // 100 - 10 - 3 = 87
		{"自定义扣分项", 0, 0, 55.5, 45},
		{"自定义扣满归零", 0, 0, 100, 0},
		{"超限夹取到100", 0, 0, -1, 100}, // 负扣分按 0 处理
		{"叠加超限夹取到0", 1, 1, 100, 0},
		{"四舍五入进位", 0.001, 0, 0, 100}, // 100-0.04=99.96 → 100
		{"四舍五入舍去", 0.05, 0, 0, 98},   // 100-2=98
		{"微小主键重复", 0, 0.001, 0, 100}, // 100-0.03=99.97 → 100
		{"半分四舍五入", 0.125, 0, 0, 95},  // 100-5=95
	}
	for _, c := range cases {
		if got := HealthScore(c.nullRate, c.dup, c.extra); got != c.want {
			t.Fatalf("%s: HealthScore(%v, %v, %v) = %d, want %d",
				c.name, c.nullRate, c.dup, c.extra, got, c.want)
		}
	}
	// 边界夹取：极端输入不越界
	if got := HealthScore(1000, 1000, 100000); got != 0 {
		t.Fatalf("极端扣分应夹取到 0, got %d", got)
	}
}

// 测试比例计算：分母为 0（空表/空单元格）返回 0，不产生 NaN
func TestRatio(t *testing.T) {
	if ratio(3, 0) != 0 || ratio(0, -5) != 0 {
		t.Fatalf("分母非正应返回 0")
	}
	if ratio(1, 4) != 0.25 {
		t.Fatalf("ratio(1,4) 应为 0.25")
	}
}

// 测试目标库名解析：请求显式指定优先，否则从 DSN 解析默认库，均无则报错
func TestProfileDbName(t *testing.T) {
	// 请求优先（即使 DSN 含库名）
	db, err := ProfileDbName("mall", "user:pass@tcp(127.0.0.1:3306)/other?charset=utf8mb4")
	if err != nil || db != "mall" {
		t.Fatalf("请求 db 优先, got %q, %v", db, err)
	}
	// DSN 解析
	db, err = ProfileDbName("", "user:pass@tcp(127.0.0.1:3306)/dwd_orders?charset=utf8mb4&parseTime=true")
	if err != nil || db != "dwd_orders" {
		t.Fatalf("应从 DSN 解析库名, got %q, %v", db, err)
	}
	// 无默认库的 DSN 且请求为空 → 报错
	if _, err := ProfileDbName("", "user:pass@tcp(127.0.0.1:3306)/"); !errors.Is(err, ErrProfileDbUnknown) {
		t.Fatalf("无库名应报 ErrProfileDbUnknown, got %v", err)
	}
	// 非法 DSN → 报错
	if _, err := ProfileDbName("", "::not-a-dsn"); !errors.Is(err, ErrProfileDbUnknown) {
		t.Fatalf("非法 DSN 应报 ErrProfileDbUnknown, got %v", err)
	}
}

// 测试逐列明细 json：数组序列化含全部字段，nil 输入输出空数组
func TestBuildColumnProfileJSON(t *testing.T) {
	cols := []*columnProfile{
		{Column: "id", NullCnt: 0, NullRate: 0, DistinctCnt: 100, SampleMin: "1", SampleMax: "100"},
		{Column: "email", NullCnt: 25, NullRate: 0.25, DistinctCnt: 75, SampleMin: "", SampleMax: "z@x.com"},
	}
	js := BuildColumnProfileJSON(cols)
	var parsed []map[string]any
	if err := json.Unmarshal([]byte(js), &parsed); err != nil {
		t.Fatalf("profile_json 应为合法 json 数组: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("应含 2 列明细, got %d", len(parsed))
	}
	for _, key := range []string{"column", "null_cnt", "null_rate", "distinct_cnt", "sample_min", "sample_max"} {
		if _, ok := parsed[0][key]; !ok {
			t.Fatalf("明细缺少字段 %q", key)
		}
	}
	if !strings.Contains(js, `"column":"email"`) {
		t.Fatalf("明细应含列名: %s", js)
	}
	// nil 输入输出空数组（而非 "null"）
	if got := BuildColumnProfileJSON(nil); got != "[]" {
		t.Fatalf("nil 输入应输出空数组, got %s", got)
	}
}

// 测试通道/级别白名单函数（model 层数据字典）
func TestChannelTypeLevelWhitelist(t *testing.T) {
	if !model.IsValidChannelType("dingtalk") || !model.IsValidChannelType("feishu") {
		t.Fatalf("dingtalk/feishu 应为合法通道类型")
	}
	if model.IsValidChannelType("email") || model.IsValidChannelType("") {
		t.Fatalf("email/空串应为非法通道类型")
	}
	if !model.IsValidAlertLevel("failed") || !model.IsValidAlertLevel("partial") {
		t.Fatalf("failed/partial 应为合法级别")
	}
	if model.IsValidAlertLevel("warn") {
		t.Fatalf("warn 应为非法级别")
	}
}
