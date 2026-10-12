// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package datasvc

import (
	"context"

	"datacenter/datasvc/rpc/dataSvcClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DsCallStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// API 调用统计
func NewDsCallStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DsCallStatsLogic {
	return &DsCallStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DsCallStats 转发 datasvc rpc 查询调用统计（成功率/耗时/TopN/日趋势）:
// days<=0 取 7(近一周)后转发; overall/top_apis/daily_calls/daily_dates 全量映射,
// 切片 make 保证空序列化为 [] 而非 null, overall 为 nil 时经共享映射返回零值结构
func (l *DsCallStatsLogic) DsCallStats(req *types.DsStatsReq) (resp *types.DsStatsResp, err error) {
	days := req.Days
	if days <= 0 {
		days = 7
	}

	reply, err := l.svcCtx.DataSvcRpc.GetCallStats(l.ctx, &dataSvcClient.StatsReq{
		Beid: req.Beid,
		Days: days,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	topApis := make([]types.DsTopApi, 0, len(reply.GetTopApis()))
	for _, item := range reply.GetTopApis() {
		topApis = append(topApis, topApiToTypes(item))
	}

	dailyCalls := make([]int64, 0, len(reply.GetDailyCalls()))
	dailyCalls = append(dailyCalls, reply.GetDailyCalls()...)

	dailyDates := make([]string, 0, len(reply.GetDailyDates()))
	dailyDates = append(dailyDates, reply.GetDailyDates()...)

	return &types.DsStatsResp{
		Overall:    statsOverallToTypes(reply.GetOverall()),
		TopApis:    topApis,
		DailyCalls: dailyCalls,
		DailyDates: dailyDates,
	}, nil
}
