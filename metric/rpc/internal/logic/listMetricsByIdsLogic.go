package logic

import (
	"context"

	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMetricsByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMetricsByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMetricsByIdsLogic {
	return &ListMetricsByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按 id 批量取指标
func (l *ListMetricsByIdsLogic) ListMetricsByIds(in *metric.MetricListByIdsReq) (*metric.MetricListReply, error) {
	// todo: add your logic here and delete this line

	return &metric.MetricListReply{}, nil
}
