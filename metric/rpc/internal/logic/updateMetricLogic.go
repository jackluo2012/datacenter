package logic

import (
	"context"

	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMetricLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMetricLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMetricLogic {
	return &UpdateMetricLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新指标
func (l *UpdateMetricLogic) UpdateMetric(in *metric.UpdateMetricReq) (*metric.IdReq, error) {
	// todo: add your logic here and delete this line

	return &metric.IdReq{}, nil
}
