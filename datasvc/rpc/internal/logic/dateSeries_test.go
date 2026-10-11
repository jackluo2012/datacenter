package logic

import (
	"reflect"
	"testing"
	"time"

	"datacenter/datasvc/model"
)

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
	//单天与上限天数
	if got := BuildDateSkeleton(start, 1); !reflect.DeepEqual(got, []string{"2026-01-30"}) {
		t.Fatalf("单天骨架不符: %v", got)
	}
	if got := BuildDateSkeleton(start, maxStatsDays); len(got) != maxStatsDays {
		t.Fatalf("90 天骨架长度应为 90, got %d", len(got))
	}
}

// 骨架补零：缺日期补 0、骨架外日期丢弃、输出与骨架等长同序
func TestFillDailySeries(t *testing.T) {
	skeleton := []string{"2026-03-01", "2026-03-02", "2026-03-03", "2026-03-04"}
	rows := []*model.ApiDailyCall{
		{Day: "2026-03-01", Calls: 5},
		{Day: "2026-03-03", Calls: 2},
		{Day: "2026-02-28", Calls: 99}, //骨架外（窗口前）丢弃
		{Day: "2026-03-05", Calls: 77}, //骨架外（窗口后）丢弃
	}
	got := FillDailySeries(skeleton, rows)
	want := []int64{5, 0, 2, 0} //缺失的 03-02/03-04 补 0
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("补零结果不符:\nwant %v\ngot  %v", want, got)
	}
	//空骨架/空数据
	if got := FillDailySeries(nil, rows); len(got) != 0 {
		t.Fatalf("空骨架应返回空切片, got %v", got)
	}
	if got := FillDailySeries(skeleton, nil); !reflect.DeepEqual(got, []int64{0, 0, 0, 0}) {
		t.Fatalf("无数据应全补 0, got %v", got)
	}
	//nil 行元素安全
	if got := FillDailySeries([]string{"2026-03-01"}, []*model.ApiDailyCall{nil}); !reflect.DeepEqual(got, []int64{0}) {
		t.Fatalf("nil 行应按 0 处理, got %v", got)
	}
}
