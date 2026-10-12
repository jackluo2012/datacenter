package logic

import (
	"math"
	"sort"
	"time"

	"datacenter/quality/model"
	"datacenter/quality/rpc/quality"
)

// 质量统计口径常量（QualityStats）：
// 窗口天数默认 7、上限 30；失败 TopN 固定取前 5
const (
	defaultStatsDays = 7
	maxStatsDays     = 30
	failTopLimit     = 5
)

// dateLayout 日趋势序列的日期格式（YYYY-MM-DD）
const dateLayout = "2006-01-02"

// NormalizeStatsDays 统计窗口天数收敛：<=0 取默认 7，超过上限取 30（纯函数，可单测）
func NormalizeStatsDays(days int) int {
	if days <= 0 {
		return defaultStatsDays
	}
	if days > maxStatsDays {
		return maxStatsDays
	}
	return days
}

// BuildDateSkeleton 生成从 startDay（含）起连续 days 天的日期骨架（YYYY-MM-DD 升序），
// 作为质量大盘日趋势的横轴：startDay 收敛到当天零点保证粒度稳定；days<=0 返回空切片（纯函数，可单测）
func BuildDateSkeleton(startDay time.Time, days int) []string {
	if days <= 0 {
		return []string{}
	}
	//收敛到当天零点，保证跨调用时日期粒度稳定（不受时分秒影响）
	day := time.Date(startDay.Year(), startDay.Month(), startDay.Day(), 0, 0, 0, 0, startDay.Location())
	dates := make([]string, 0, days)
	for i := 0; i < days; i++ {
		dates = append(dates, day.AddDate(0, 0, i).Format(dateLayout))
	}
	return dates
}

// CalcPassRate 通过率（百分数 0-100，保留 2 位小数）：passed/total*100；
// 窗口内无结果（total<=0）时为 0，避免除零（纯函数，可单测）
func CalcPassRate(passed, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return math.Round(float64(passed)/float64(total)*10000) / 100
}

// TopFailRules 失败 TopN 排序兜底（纯函数，可单测）：
// SQL 已按 fail_cnt desc 排序，此处再排一次作为契约兜底——失败数降序、并列按 rule_id 升序，截取前 n；
// 输入为 nil/空返回空切片，不修改入参切片
func TopFailRules(rows []*model.QualityFailTopRow, n int) []*model.QualityFailTopRow {
	if len(rows) == 0 || n <= 0 {
		return []*model.QualityFailTopRow{}
	}
	//拷贝后排序，避免改动 model 返回的原切片
	sorted := make([]*model.QualityFailTopRow, len(rows))
	copy(sorted, rows)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i] == nil {
			return false
		}
		if sorted[j] == nil {
			return true
		}
		if sorted[i].FailCnt != sorted[j].FailCnt {
			return sorted[i].FailCnt > sorted[j].FailCnt
		}
		return sorted[i].RuleId < sorted[j].RuleId
	})
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}

// FillDailyQuality 把逐日聚合行回填到日期骨架（纯函数，可单测）：
// 输出与骨架等长同序，缺数日期补零、骨架外日期丢弃，保证大盘折线横轴连续
func FillDailyQuality(skeleton []string, rows []*model.QualityDailyRow) []*quality.DailyQualityStat {
	byDay := make(map[string]*model.QualityDailyRow, len(rows))
	for _, r := range rows {
		if r != nil {
			byDay[r.Day] = r
		}
	}
	daily := make([]*quality.DailyQualityStat, 0, len(skeleton))
	for _, d := range skeleton {
		point := &quality.DailyQualityStat{Date: d}
		if row, ok := byDay[d]; ok {
			point.Runs = row.Runs
			point.Passed = row.Passed
			point.Failed = row.Failed
		}
		daily = append(daily, point)
	}
	return daily
}

// FillHealthTrend 把逐日平均健康分行回填到日期骨架（纯函数，可单测）：
// 输出与骨架等长同序，无剖析记录的日期补 0（前端以 runs/daily 区分有无活动，趋势轴保持连续）
func FillHealthTrend(skeleton []string, rows []*model.QualityHealthRow) []*quality.HealthTrendPoint {
	byDay := make(map[string]float64, len(rows))
	for _, r := range rows {
		if r != nil {
			byDay[r.Day] = r.AvgHealth
		}
	}
	trend := make([]*quality.HealthTrendPoint, 0, len(skeleton))
	for _, d := range skeleton {
		trend = append(trend, &quality.HealthTrendPoint{
			Date:           d,
			AvgHealthScore: byDay[d], //map 缺省即 0，天然补零
		})
	}
	return trend
}
