package logic

import (
	"reflect"
	"testing"

	"datacenter/ai/rpc/ai"
	"datacenter/metadata/rpc/metadataClient"
	"datacenter/metric/rpc/metricsvc"
)

// 测试候选打分：name(3) > comment(2) > detail(1)，命中数累加计分
func TestScoreCandidate(t *testing.T) {
	keywords := []string{"订单", "gmv"}

	cases := []struct {
		name      string
		candidate *ai.Candidate
		wantScore int
		wantHits  int
	}{
		{
			name:      "name 命中一个关键词",
			candidate: &ai.Candidate{Name: "订单表"},
			wantScore: 3,
			wantHits:  1,
		},
		{
			name:      "comment 命中优先级低于 name",
			candidate: &ai.Candidate{Name: "用户表", Comment: "含订单信息"},
			wantScore: 2,
			wantHits:  1,
		},
		{
			name:      "detail(fqn) 命中优先级最低",
			candidate: &ai.Candidate{Name: "支付表", Comment: "支付流水", Detail: "mysql://mall.gmv_fact"},
			wantScore: 1,
			wantHits:  1,
		},
		{
			name:      "多关键词命中累加（name+name）",
			candidate: &ai.Candidate{Name: "gmv订单视图"},
			wantScore: 6,
			wantHits:  2,
		},
		{
			name:      "大小写不敏感",
			candidate: &ai.Candidate{Name: "GMV指标"},
			wantScore: 3,
			wantHits:  1,
		},
		{
			name:      "全不命中",
			candidate: &ai.Candidate{Name: "用户表", Comment: "用户主表", Detail: "mysql://mall.users"},
			wantScore: 0,
			wantHits:  0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			score, hits := scoreCandidate(c.candidate, keywords)
			if score != c.wantScore || hits != c.wantHits {
				t.Errorf("scoreCandidate = (%d,%d), 期望 (%d,%d)", score, hits, c.wantScore, c.wantHits)
			}
		})
	}
}

// 测试排序与截取：分数降序，同分按命中数降序，同分同命中保持稳定
func TestRankCandidates(t *testing.T) {
	a := &ai.Candidate{Type: "entity", Id: 1, Name: "订单表", Detail: "mysql://mall.orders"}        //name 命中"订单"：3 分 1 命中
	b := &ai.Candidate{Type: "metric", Id: 2, Name: "订单gmv", Comment: "成交金额"}                    //name 命中两词：6 分 2 命中
	c := &ai.Candidate{Type: "metric", Id: 3, Name: "统计视图", Comment: "订单相关", Detail: "gmv expr"} //comment(2)+detail(1)：3 分 2 命中
	d := &ai.Candidate{Type: "entity", Id: 4, Name: "无关联"}                                       //0 分
	in := []*ai.Candidate{a, b, c, d}
	got := RankCandidates(in, []string{"订单", "gmv"}, 10)

	//b(6) > c(3分2命中) > a(3分1命中) > d(0)
	want := []*ai.Candidate{b, c, a, d}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("RankCandidates 顺序 = %v, 期望 %v", candIds(got), candIds(want))
	}

	//截取 top n
	got2 := RankCandidates(in, []string{"订单", "gmv"}, 2)
	if len(got2) != 2 || got2[0] != b || got2[1] != c {
		t.Errorf("RankCandidates top2 = %v, 期望 [b c]", candIds(got2))
	}

	//空入参
	if got3 := RankCandidates(nil, []string{"x"}, 5); got3 != nil {
		t.Errorf("RankCandidates(nil) = %v, 期望 nil", got3)
	}
}

func candIds(cands []*ai.Candidate) []int64 {
	ids := make([]int64, 0, len(cands))
	for _, c := range cands {
		ids = append(ids, c.Id)
	}
	return ids
}

// 测试实体/指标候选转换：entity→detail=fqn，metric→agg(biz_field)/expr，去重保序
func TestBuildAndDedupeCandidates(t *testing.T) {
	entities := []*metadataClient.MetaEntity{
		{Id: 1, Name: "订单表", Fqn: "mysql://mall.orders", Comment: "订单主表"},
		{Id: 1, Name: "订单表", Fqn: "mysql://mall.orders", Comment: "订单主表"}, //与上一条重复（多关键词命中）
	}
	ms := []*metricsvc.Metric{
		{Id: 2, NameCn: "GMV", NameEn: "gmv", AggType: "sum", BizField: "amount"},
		{Id: 3, NameCn: "", NameEn: "aov", Expr: "gmv/user_cnt"}, //无中文名回退英文名
	}

	got := buildCandidates(entities, ms)
	want := []*ai.Candidate{
		{Type: "entity", Id: 1, Name: "订单表", Detail: "mysql://mall.orders", Comment: "订单主表"},
		{Type: "metric", Id: 2, Name: "GMV", Detail: "sum(amount)"},
		{Type: "metric", Id: 3, Name: "aov", Detail: "gmv/user_cnt"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("候选合并结果 = %v, 期望 %v", got, want)
	}
}
