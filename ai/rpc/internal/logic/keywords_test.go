package logic

import (
	"reflect"
	"testing"
)

// 测试停用词剔除 + 整句/分词提取
func TestExtractKeywords(t *testing.T) {
	cases := []struct {
		name     string
		question string
		want     []string
	}{
		{
			name:     "空问句",
			question: "   ",
			want:     nil,
		},
		{
			name:     "纯中文停用词剔除",
			question: "查询订单表有多少数据",
			//整句在前，短中文块整块保留（<4 不切 bigram）
			want: []string{"订单表数据", "订单表", "数据"},
		},
		{
			name:     "单字虚词剔除",
			question: "用户的订单数量",
			want:     []string{"用户订单", "用户", "订单"},
		},
		{
			name:     "中英混排",
			question: "查询orders表的gmv",
			//整句拼接不补空格；单字中文块"表"被过滤
			want: []string{"orders表gmv", "orders", "gmv"},
		},
		{
			name:     "英文大写统一小写",
			question: "查看GMV指标",
			want:     []string{"gmv指标", "gmv", "指标"},
		},
		{
			name:     "长中文块切bigram",
			question: "订单明细",
			want:     []string{"订单明细", "订单", "单明", "明细"},
		},
		{
			name:     "去重保序",
			question: "订单订单",
			want:     []string{"订单订单", "订单", "单订"},
		},
		{
			name:     "全停用词兜底原问句",
			question: "是的",
			want:     []string{"是的"},
		},
		{
			name:     "标点分隔",
			question: "统计一下：用户表，行数？",
			want:     []string{"用户表行数", "用户表", "行数"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ExtractKeywords(c.question)
			if c.want == nil {
				if got != nil {
					t.Errorf("ExtractKeywords(%q) = %v, 期望 nil", c.question, got)
				}
				return
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("ExtractKeywords(%q) = %v, 期望 %v", c.question, got, c.want)
			}
		})
	}
}

// 测试多字停用词最长优先剔除："有多少" 先于 "有"/"少" 处理
func TestRemoveStopWordsLongestFirst(t *testing.T) {
	if got := removeStopWords("有多少用户"); got != "用户" {
		t.Errorf("removeStopWords = %q, 期望 %q", got, "用户")
	}
}
