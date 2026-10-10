package logic

import (
	"context"

	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMetricsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMetricsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMetricsLogic {
	return &ListMetricsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 指标列表：分页 + metric_type 过滤 + keyword 对 name_en/name_cn 模糊匹配
func (l *ListMetricsLogic) ListMetrics(in *metric.MetricListReq) (*metric.MetricListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, total, err := l.svcCtx.MetricModel.FindList(l.ctx, in.Beid, in.MetricType, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	return &metric.MetricListReply{List: metricListToPb(list), Total: total}, nil
}
