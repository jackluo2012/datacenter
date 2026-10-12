package logic

import (
	"context"
	"time"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type QualityStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQualityStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QualityStatsLogic {
	return &QualityStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 质量统计大盘（供前端质量可视化：通过率趋势/失败 TopN/健康分趋势）：
// 一次 Logic 内 4 条聚合 SQL 完成（总览/逐日/失败 TopN/健康分趋势），
// 逐日序列经 Go 侧日期骨架补零，保证近 N 天横轴连续。
// 口径：days 默认 7、上限 30；窗口 [近 N 天零点, 明日零点)；
// RunJob 次数按结果表 job+日去重计数；通过率为窗口内结果行 pass_flag=1 占比（百分数 0-100）
func (l *QualityStatsLogic) QualityStats(in *quality.StatsReq) (*quality.StatsReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	days := NormalizeStatsDays(int(in.Days))

	//统计窗口：[近 days 天的零点, 明日零点)，与日趋势日期骨架严格对齐
	today := time.Now()
	startDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()).
		AddDate(0, 0, -(days - 1))
	end := startDay.AddDate(0, 0, days)

	//1. 总览：一次 SQL 聚合（RunJob 次数/结果总数/通过数）
	overview, err := l.svcCtx.StatsModel.Overview(l.ctx, in.Beid, startDay, end)
	if err != nil {
		return nil, err
	}
	passRate := CalcPassRate(overview.PassedCnt, overview.TotalCnt)

	//2. 逐日统计：按 run_at 天数归组 + 骨架补零
	dailyRows, err := l.svcCtx.StatsModel.DailyResults(l.ctx, in.Beid, startDay, end)
	if err != nil {
		return nil, err
	}
	skeleton := BuildDateSkeleton(startDay, days)
	daily := FillDailyQuality(skeleton, dailyRows)

	//3. 失败 TopN：失败次数前 5（model 层 join 规则表出名字，排序兜底在纯函数）
	failRows, err := l.svcCtx.StatsModel.FailTopRules(l.ctx, in.Beid, startDay, end, failTopLimit)
	if err != nil {
		return nil, err
	}
	topRows := TopFailRules(failRows, failTopLimit)
	failTops := make([]*quality.FailTopRule, 0, len(topRows))
	for _, row := range topRows {
		failTops = append(failTops, &quality.FailTopRule{
			RuleId:   row.RuleId,
			RuleName: row.RuleName,
			FailCnt:  row.FailCnt,
		})
	}

	//4. 健康分趋势：剖析表按日平均 + 骨架补零
	healthRows, err := l.svcCtx.StatsModel.HealthTrend(l.ctx, in.Beid, startDay, end)
	if err != nil {
		return nil, err
	}
	healthTrend := FillHealthTrend(skeleton, healthRows)

	return &quality.StatsReply{
		TotalRuns:    overview.TotalRuns,
		PassRate:     passRate,
		Daily:        daily,
		FailTopRules: failTops,
		HealthTrend:  healthTrend,
	}, nil
}
