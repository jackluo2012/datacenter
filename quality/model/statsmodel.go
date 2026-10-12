package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 质量统计聚合（供 QualityStats 一次 Logic 内 4 条聚合 SQL 完成）：
// dc_quality_result 无 beid 列，多租户过滤统一经 join dc_quality_job（任务表带 beid）实现；
// date 归组统一用 date_format 输出 YYYY-MM-DD 字符串——语义等价按 date() 归组，
// 同时规避 DSN parseTime=true 下 DATE/DATETIME 列扫描为 time.Time 的类型漂移（datasvc callstats 先例）。
const (
	// statsResultJoinJob 结果表 join 任务表骨架（多租户 beid 过滤 + 时间窗口）
	statsResultJoinJob = "`dc_quality_result` r join `dc_quality_job` j on j.`id` = r.`job_id`"
	// statsResultWhere 结果表统计公共过滤：窗口 [begin, end) + beid
	statsResultWhere = "where r.`run_at` >= ? and r.`run_at` < ? and j.`beid` = ?"
	// statsPassedExpr 通过数表达式（pass_flag=1 计数，空窗口 ifnull 兜底为 0）
	statsPassedExpr = "ifnull(sum(r.`pass_flag`), 0)"
)

// 质量统计聚合行（扫描列与 select 显式列清单一一对应，规避少列扫宽 struct 报错）
type (
	// QualityOverview 窗口总览：RunJob 次数（job+日去重）与结果通过/总数
	QualityOverview struct {
		TotalRuns int64 `db:"total_runs"` // RunJob 次数（job_id+日 去重）
		TotalCnt  int64 `db:"total_cnt"`  // 检查结果总行数
		PassedCnt int64 `db:"passed_cnt"` // 通过结果行数（pass_flag=1）
	}

	// QualityDailyRow 逐日聚合行（day 为 YYYY-MM-DD 字符串）
	QualityDailyRow struct {
		Day    string `db:"day"`    // 归组日期 YYYY-MM-DD
		Runs   int64  `db:"runs"`   // 当日 RunJob 次数（job 去重）
		Passed int64  `db:"passed"` // 当日通过结果数
		Failed int64  `db:"failed"` // 当日不通过结果数
	}

	// QualityFailTopRow 失败 TopN 行（join 规则表出名字，规则已删除时 name 为空串）
	QualityFailTopRow struct {
		RuleId   int64  `db:"rule_id"`
		RuleName string `db:"rule_name"`
		FailCnt  int64  `db:"fail_cnt"`
	}

	// QualityHealthRow 逐日平均健康分行
	QualityHealthRow struct {
		Day       string  `db:"day"`        // 归组日期 YYYY-MM-DD
		AvgHealth float64 `db:"avg_health"` // 当日平均健康分 0-100
	}
)

// StatsModel 质量统计聚合接口
type StatsModel interface {
	// Overview 窗口总览：近 N 天 RunJob 次数（job+日去重）、结果总数、通过数
	Overview(ctx context.Context, beid int64, begin, end time.Time) (*QualityOverview, error)
	// DailyResults 窗口内逐日统计（按 run_at 天数归组），升序
	DailyResults(ctx context.Context, beid int64, begin, end time.Time) ([]*QualityDailyRow, error)
	// FailTopRules 窗口内失败次数 TopN（join dc_quality_rule 出名字），按失败数倒序
	FailTopRules(ctx context.Context, beid int64, begin, end time.Time, limit int64) ([]*QualityFailTopRow, error)
	// HealthTrend 窗口内逐日平均健康分（dc_quality_profile 自带 beid，按 ran_at 天数归组），升序
	HealthTrend(ctx context.Context, beid int64, begin, end time.Time) ([]*QualityHealthRow, error)
}

type defaultStatsModel struct {
	conn sqlx.SqlConn
}

// NewStatsModel 创建质量统计聚合 model（与业务 model 共用中台主库连接）
func NewStatsModel(conn sqlx.SqlConn) StatsModel {
	return &defaultStatsModel{conn: conn}
}

// Overview 窗口总览：一次 SQL 聚合——
// total_runs 用 count(distinct job_id, 日期串) 按「任务+日」去重近似 RunJob 次数
// （同一天同任务多次运行只计 1 次；join 任务表为多对一不放大行数，count(*) 仍为结果行数）
func (m *defaultStatsModel) Overview(ctx context.Context, beid int64, begin, end time.Time) (*QualityOverview, error) {
	query := fmt.Sprintf(
		"select ifnull(count(distinct r.`job_id`, date_format(r.`run_at`, '%%Y-%%m-%%d')), 0) as `total_runs`, "+
			"count(*) as `total_cnt`, %s as `passed_cnt` "+
			"from %s %s", statsPassedExpr, statsResultJoinJob, statsResultWhere)
	var resp QualityOverview
	if err := m.conn.QueryRowCtx(ctx, &resp, query, begin, end, beid); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DailyResults 逐日统计：runs 当日任务去重数、passed 通过数、failed=总数-通过数（select/group by
// 均用 date_format 输出 YYYY-MM-DD，规避 parseTime 类型漂移）
func (m *defaultStatsModel) DailyResults(ctx context.Context, beid int64, begin, end time.Time) ([]*QualityDailyRow, error) {
	query := fmt.Sprintf(
		"select date_format(r.`run_at`, '%%Y-%%m-%%d') as `day`, "+
			"count(distinct r.`job_id`) as `runs`, %s as `passed`, "+
			"count(*) - %s as `failed` "+
			"from %s %s group by `day` order by `day` asc",
		statsPassedExpr, statsPassedExpr, statsResultJoinJob, statsResultWhere)
	var list []*QualityDailyRow
	if err := m.conn.QueryRowsCtx(ctx, &list, query, begin, end, beid); err != nil {
		return nil, err
	}
	return list, nil
}

// FailTopRules 失败次数 TopN：仅统计 pass_flag=0，left join 规则表补名字
// （规则已删除不丢失败记录，名字兜底空串），同失败数按 rule_id 升序保证稳定排序
func (m *defaultStatsModel) FailTopRules(ctx context.Context, beid int64, begin, end time.Time,
	limit int64) ([]*QualityFailTopRow, error) {
	// 窗口 + 多租户过滤之上追加「仅不通过」条件
	failWhere := statsResultWhere + " and r.`pass_flag` = 0"
	query := fmt.Sprintf(
		"select r.`rule_id` as `rule_id`, ifnull(ru.`name`, '') as `rule_name`, count(*) as `fail_cnt` "+
			"from %s left join `dc_quality_rule` ru on ru.`id` = r.`rule_id` %s "+
			"group by r.`rule_id`, ru.`name` "+
			"order by `fail_cnt` desc, r.`rule_id` asc limit ?", statsResultJoinJob, failWhere)
	var list []*QualityFailTopRow
	if err := m.conn.QueryRowsCtx(ctx, &list, query, begin, end, beid, limit); err != nil {
		return nil, err
	}
	return list, nil
}

// HealthTrend 健康分趋势：dc_quality_profile 按日平均健康分（表自带 beid 无需 join），
// avg 结果 round 保留 2 位小数
func (m *defaultStatsModel) HealthTrend(ctx context.Context, beid int64, begin, end time.Time) ([]*QualityHealthRow, error) {
	query := "select date_format(p.`ran_at`, '%Y-%m-%d') as `day`, " +
		"round(ifnull(avg(p.`health_score`), 0), 2) as `avg_health` " +
		"from `dc_quality_profile` p " +
		"where p.`ran_at` >= ? and p.`ran_at` < ? and p.`beid` = ? " +
		"group by `day` order by `day` asc"
	var list []*QualityHealthRow
	if err := m.conn.QueryRowsCtx(ctx, &list, query, begin, end, beid); err != nil {
		return nil, err
	}
	return list, nil
}
