package logic

import (
	"context"
	"errors"

	"datacenter/metric/model"
	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMetricLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMetricLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMetricLogic {
	return &GetMetricLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 指标详情
func (l *GetMetricLogic) GetMetric(in *metric.IdReq) (*metric.MetricReply, error) {
	if in.Id <= 0 {
		return nil, ErrMetricNotFound
	}
	m, err := l.svcCtx.MetricModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrMetricNotFound
		}
		return nil, err
	}
	return &metric.MetricReply{Metric: metricToPb(m)}, nil
}
