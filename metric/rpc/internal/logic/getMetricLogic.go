package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &metric.MetricReply{}, nil
}
