package logic

import (
	"reflect"
	"testing"
	"time"

	"datacenter/quality/model"
)

// 窗口天数收敛：缺省 7 / 上限 30 / 原样透传
func TestNormalizeStatsDays(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{0, defaultStatsDays},  //缺省取 7
		{-5, defaultStatsDays}, //负数取 7
		{1, 1},                 //合法下界原样
		{30, 30},               //合法上界原样
		{31, maxStatsDays},     //超上限取 30
		{999, maxStatsDays},    //远超上限取 30
	}
	for _, c := range cases {
		if got := NormalizeStatsDays(c.in); got != c.want {
			t.Fatalf("NormalizeStatsDays(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

// 日期骨架：连续性 / 跨月跨年 / 边界 days<=0 / 天数上限形态
func TestBuildDateSkeleton(t *testing.T) {
	start := time.Date(2026, 1, 30, 15, 42, 9, 0, time.UTC)
	//跨月连续 4 天（时分秒应被收敛掉）
	got := BuildDateSkeleton(start, 4)
	want := []string{"2026-01-30", "2026-01-31", "2026-02-01", "2026-02-02"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("跨月骨架不符:\nwant %v\ngot  %v", want, got)
	}
	//跨年
	got = BuildDateSkeleton(time.Date(2026, 12, 30, 0, 0, 0, 0, time.UTC), 3)
	want = []string{"2026-12-30", "2026-12-31", "2027-01-01"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("跨年骨架不符: %v", got)
	}
	//days<=0 返回空切片
	if got := BuildDateSkeleton(start, 0); len(got) != 0 {
		t.Fatalf("days=0 应返回空切片, got %v", got)
	}
	if got := BuildDateSkeleton(start, -3); len(got) != 0 {
		t.Fatalf("days<0 应返回空切片, got %v", got)
	}
	//单天与默认/上限天数形态
	if got := BuildDateSkeleton(start, 1); !reflect.DeepEqual(got, []string{"2026-01-30"}) {
		t.Fatalf("单天骨架不符: %v", got)
	}
	if got := BuildDateSkeleton(start, defaultStatsDays); len(got) != defaultStatsDays {
		t.Fatalf("默认窗口骨架长度应为 7, got %d", len(got))
	}
	if got := BuildDateSkeleton(start, maxStatsDays); len(got) != maxStatsDays {
		t.Fatalf("上限窗口骨架长度应为 30, got %d", len(got))
	}
	//闰年 2 月 29 日
	got = BuildDateSkeleton(time.Date(2028, 2, 28, 0, 0, 0, 0, time.UTC), 3)
	want = []string{"2028-02-28", "2028-02-29", "2028-03-01"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("闰年骨架不符: %v", got)
	}
}

// 通过率计算：0 除保护 / 常规比例 / 百分数换算 / 2 位小数舍入
func TestCalcPassRate(t *testing.T) {
	cases := []struct {
		passed, total int64
		want          float64
	}{
		{0, 0, 0},        //空窗口除零保护
		{10, 0, 0},       //无总数兜底
		{0, 10, 0},       //全不通过
		{10, 10, 100},    //全通过
		{3, 4, 75},       //常规比例
		{1, 3, 33.33},    //循环小数截到 2 位
		{2, 3, 66.67},    //四舍五入进位
		{1, 6, 16.67},    //16.666... -> 16.67
		{1, 8, 12.5},     //精确 1 位小数
		{199, 200, 99.5}, //99.5 不再进位为 100
	}
	for _, c := range cases {
		if got := CalcPassRate(c.passed, c.total); got != c.want {
			t.Fatalf("CalcPassRate(%d,%d) = %v, want %v", c.passed, c.total, got, c.want)
		}
	}
}

// 失败 TopN 排序兜底：降序 / 并列按 rule_id 升序 / 截取前 n / 空与 nil 安全 / 不改入参
func TestTopFailRules(t *testing.T) {
	rows := []*model.QualityFailTopRow{
		{RuleId: 7, RuleName: "g", FailCnt: 2},
		{RuleId: 3, RuleName: "c", FailCnt: 9},
		{RuleId: 5, RuleName: "e", FailCnt: 9}, //与 rule 3 并列，id 升序应排前
		{RuleId: 1, RuleName: "a", FailCnt: 12},
		{RuleId: 9, RuleName: "i", FailCnt: 1},
		{RuleId: 4, RuleName: "d", FailCnt: 4},
	}
	got := TopFailRules(rows, failTopLimit)
	wantIds := []int64{1, 3, 5, 4, 7} //12,9,9,4,2（并列 3<5），第 6 名 id=9 被截断
	if len(got) != failTopLimit {
		t.Fatalf("TopN 应截取前 %d, got %d", failTopLimit, len(got))
	}
	for i, w := range wantIds {
		if got[i].RuleId != w {
			t.Fatalf("TopN 第 %d 位应为 rule %d, got %d", i, w, got[i].RuleId)
		}
	}
	//入参顺序未被修改（纯函数）
	if rows[0].RuleId != 7 || rows[2].RuleId != 5 {
		t.Fatalf("入参切片不应被重排: %v", rows)
	}
	//不足 n 原样全量返回（有序）
	got = TopFailRules(rows[:2], failTopLimit)
	if len(got) != 2 || got[0].RuleId != 3 || got[1].RuleId != 7 {
		t.Fatalf("不足 n 应全量有序返回: %v", got)
	}
	//空/nil/n<=0 返回空切片
	if got := TopFailRules(nil, failTopLimit); len(got) != 0 {
		t.Fatalf("nil 输入应返回空切片, got %v", got)
	}
	if got := TopFailRules([]*model.QualityFailTopRow{}, failTopLimit); len(got) != 0 {
		t.Fatalf("空输入应返回空切片, got %v", got)
	}
	if got := TopFailRules(rows, 0); len(got) != 0 {
		t.Fatalf("n=0 应返回空切片, got %v", got)
	}
	//nil 行元素不 panic 且沉底
	withNil := []*model.QualityFailTopRow{nil, {RuleId: 2, RuleName: "b", FailCnt: 3}}
	got = TopFailRules(withNil, failTopLimit)
	if len(got) != 2 || got[0] == nil || got[0].RuleId != 2 {
		t.Fatalf("nil 行应沉底: %v", got)
	}
}

// 逐日统计回填骨架：缺日补零 / 骨架外丢弃 / 输出与骨架等长同序
func TestFillDailyQuality(t *testing.T) {
	skeleton := []string{"2026-03-01", "2026-03-02", "2026-03-03", "2026-03-04"}
	rows := []*model.QualityDailyRow{
		{Day: "2026-03-01", Runs: 2, Passed: 8, Failed: 1},
		{Day: "2026-03-03", Runs: 1, Passed: 3, Failed: 5},
		{Day: "2026-02-28", Runs: 9, Passed: 9, Failed: 9}, //窗口前丢弃
		{Day: "2026-03-05", Runs: 9, Passed: 9, Failed: 9}, //窗口后丢弃
	}
	got := FillDailyQuality(skeleton, rows)
	if len(got) != len(skeleton) {
		t.Fatalf("输出应与骨架等长: %d", len(got))
	}
	//缺数日期补零且日期同序
	want := []struct {
		date                 string
		runs, passed, failed int64
	}{
		{"2026-03-01", 2, 8, 1},
		{"2026-03-02", 0, 0, 0},
		{"2026-03-03", 1, 3, 5},
		{"2026-03-04", 0, 0, 0},
	}
	for i, w := range want {
		if got[i].Date != w.date || got[i].Runs != w.runs ||
			got[i].Passed != w.passed || got[i].Failed != w.failed {
			t.Fatalf("第 %d 日不符: want %+v, got %+v", i, w, got[i])
		}
	}
	//空骨架/空数据
	if got := FillDailyQuality(nil, rows); len(got) != 0 {
		t.Fatalf("空骨架应返回空切片, got %v", got)
	}
	if got := FillDailyQuality(skeleton, nil); len(got) != 4 {
		t.Fatalf("无数据应返回等长补零序列, got %d", len(got))
	}
	//nil 行元素安全
	if got := FillDailyQuality(skeleton[:1], []*model.QualityDailyRow{nil}); got[0].Runs != 0 {
		t.Fatalf("nil 行应按 0 处理, got %+v", got[0])
	}
}

// 健康分趋势回填骨架：缺日补 0 / 骨架外丢弃
func TestFillHealthTrend(t *testing.T) {
	skeleton := []string{"2026-03-01", "2026-03-02", "2026-03-03"}
	rows := []*model.QualityHealthRow{
		{Day: "2026-03-01", AvgHealth: 90.5},
		{Day: "2026-03-03", AvgHealth: 88},
		{Day: "2026-02-28", AvgHealth: 70}, //窗口外丢弃
	}
	got := FillHealthTrend(skeleton, rows)
	if len(got) != 3 {
		t.Fatalf("输出应与骨架等长: %d", len(got))
	}
	if got[0].Date != "2026-03-01" || got[0].AvgHealthScore != 90.5 {
		t.Fatalf("第 1 日不符: %+v", got[0])
	}
	if got[1].AvgHealthScore != 0 {
		t.Fatalf("缺剖析日期应补 0: %+v", got[1])
	}
	if got[2].AvgHealthScore != 88 {
		t.Fatalf("第 3 日不符: %+v", got[2])
	}
}
