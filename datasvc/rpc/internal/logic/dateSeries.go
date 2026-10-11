package logic

import (
	"time"

	"datacenter/datasvc/model"
)

// dateLayout 日趋势序列的日期格式（YYYY-MM-DD）
const dateLayout = "2006-01-02"

// BuildDateSkeleton 生成从 startDay（含）起连续 days 天的日期骨架（YYYY-MM-DD 升序），
// 作为日趋势的横轴：days<=0 返回空切片（纯函数，可单测）
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

// FillDailySeries 把按天聚合结果回填到日期骨架（纯函数）：
// 骨架内无数据的日期补 0，骨架外的日期丢弃，输出与骨架等长且顺序一致
func FillDailySeries(skeleton []string, rows []*model.ApiDailyCall) []int64 {
	byDay := make(map[string]int64, len(rows))
	for _, r := range rows {
		if r != nil {
			byDay[r.Day] = r.Calls
		}
	}
	calls := make([]int64, len(skeleton))
	for i, d := range skeleton {
		calls[i] = byDay[d] //map 缺省即 0，天然补零
	}
	return calls
}
