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

// 指标列表
func (l *ListMetricsLogic) ListMetrics(in *metric.MetricListReq) (*metric.MetricListReply, error) {
	// todo: add your logic here and delete this line

	return &metric.MetricListReply{}, nil
}
