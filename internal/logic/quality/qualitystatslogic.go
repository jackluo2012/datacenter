// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package quality

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/quality/rpc/qualityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QualityStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 质量统计大盘
func NewQualityStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QualityStatsLogic {
	return &QualityStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QualityStats 转发 quality rpc 质量统计大盘（近 N 天：总览/逐日/失败 TopN/健康分趋势）:
// beid/days 原样透传（days 缺省 7、上限 30 的收敛在 rpc 侧统一处理）;
// 切片 make 保证空序列化为 [] 而非 null
func (l *QualityStatsLogic) QualityStats(req *types.StatsQReq) (resp *types.QualityStatsResp, err error) {
	reply, err := l.svcCtx.QualityRpc.QualityStats(l.ctx, &qualityClient.StatsReq{
		Beid: req.Beid,
		Days: int32(req.Days),
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}

	daily := make([]types.QualityDailyStat, 0, len(reply.GetDaily()))
	for _, item := range reply.GetDaily() {
		daily = append(daily, types.QualityDailyStat{
			Date:   item.GetDate(),
			Runs:   item.GetRuns(),
			Passed: item.GetPassed(),
			Failed: item.GetFailed(),
		})
	}

	failTopRules := make([]types.QualityFailTopRule, 0, len(reply.GetFailTopRules()))
	for _, item := range reply.GetFailTopRules() {
		failTopRules = append(failTopRules, types.QualityFailTopRule{
			RuleId:   item.GetRuleId(),
			RuleName: item.GetRuleName(),
			FailCnt:  item.GetFailCnt(),
		})
	}

	healthTrend := make([]types.QualityHealthPoint, 0, len(reply.GetHealthTrend()))
	for _, item := range reply.GetHealthTrend() {
		healthTrend = append(healthTrend, types.QualityHealthPoint{
			Date:           item.GetDate(),
			AvgHealthScore: item.GetAvgHealthScore(),
		})
	}

	return &types.QualityStatsResp{
		TotalRuns:    reply.GetTotalRuns(),
		PassRate:     reply.GetPassRate(),
		Daily:        daily,
		FailTopRules: failTopRules,
		HealthTrend:  healthTrend,
	}, nil
}
