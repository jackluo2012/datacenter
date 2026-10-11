package logic

import (
	"context"
	"math"
	"time"

	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// 统计口径约束：近 N 天，默认 7、上限 90；TopN 固定取调用量前 10
const (
	defaultStatsDays = 7
	maxStatsDays     = 90
	topApiLimit      = 10
)

type GetCallStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCallStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCallStatsLogic {
	return &GetCallStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 调用统计：总览一次 SQL 聚合（总调用/ok/error/denied/平均耗时/cache 命中）
// + TopN API（调用量前 10，join dc_api 出 path）+ 日趋势（Go 侧日期骨架补零）
func (l *GetCallStatsLogic) GetCallStats(in *datasvc.StatsReq) (*datasvc.StatsReply, error) {
	//窗口天数：默认 7，上限 90
	days := int(in.Days)
	if days <= 0 {
		days = defaultStatsDays
	}
	if days > maxStatsDays {
		days = maxStatsDays
	}
	//统计窗口：[近 days 天的零点, 明日零点)，与日趋势日期序列严格对齐
	today := time.Now()
	startDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()).
		AddDate(0, 0, -(days - 1))
	end := startDay.AddDate(0, 0, days)

	//1. 总览：一次 SQL 聚合
	overall, err := l.svcCtx.CallLogModel.Stats(in.Beid, startDay, end)
	if err != nil {
		return nil, err
	}
	//成功率 = ok / total（窗口内无调用时为 0），保留 4 位小数
	successRate := 0.0
	if overall.TotalCalls > 0 {
		successRate = math.Round(float64(overall.OkCalls)/float64(overall.TotalCalls)*10000) / 10000
	}

	//2. TopN API：按调用量前 10（path 由 model 层 join dc_api 所得）
	tops, err := l.svcCtx.CallLogModel.TopApis(in.Beid, startDay, end, topApiLimit)
	if err != nil {
		return nil, err
	}
	topApis := make([]*datasvc.TopApi, 0, len(tops))
	for _, tp := range tops {
		topApis = append(topApis, &datasvc.TopApi{
			ApiId:     tp.ApiId,
			Path:      tp.Path,
			Calls:     tp.Calls,
			AvgCostMs: tp.AvgCostMs,
		})
	}

	//3. 日趋势：按天聚合 + Go 侧生成连续日期骨架并补零
	rows, err := l.svcCtx.CallLogModel.DailyCalls(in.Beid, startDay, end)
	if err != nil {
		return nil, err
	}
	dates := BuildDateSkeleton(startDay, days)
	dailyCalls := FillDailySeries(dates, rows)

	return &datasvc.StatsReply{
		Overall: &datasvc.ApiCallStats{
			TotalCalls:  overall.TotalCalls,
			OkCalls:     overall.OkCalls,
			ErrorCalls:  overall.ErrorCalls,
			DeniedCalls: overall.DeniedCalls,
			SuccessRate: successRate,
			AvgCostMs:   overall.AvgCostMs,
			CacheHits:   overall.CacheHits,
		},
		TopApis:    topApis,
		DailyCalls: dailyCalls,
		DailyDates: dates,
	}, nil
}
