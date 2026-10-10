package logic

import (
	"context"

	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMetricLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMetricLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMetricLogic {
	return &DeleteMetricLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除指标
func (l *DeleteMetricLogic) DeleteMetric(in *metric.IdReq) (*metric.IdReq, error) {
	// todo: add your logic here and delete this line

	return &metric.IdReq{}, nil
}
