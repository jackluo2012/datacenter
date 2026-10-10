package logic

import (
	"context"

	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMetricLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateMetricLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMetricLogic {
	return &CreateMetricLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建指标
func (l *CreateMetricLogic) CreateMetric(in *metric.CreateMetricReq) (*metric.MetricReply, error) {
	// todo: add your logic here and delete this line

	return &metric.MetricReply{}, nil
}
