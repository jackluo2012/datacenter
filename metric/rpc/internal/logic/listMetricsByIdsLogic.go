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

// 按 id 批量取指标（model 内自动去重并分片，total 返回实际命中条数）
func (l *ListMetricsByIdsLogic) ListMetricsByIds(in *metric.MetricListByIdsReq) (*metric.MetricListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, err := l.svcCtx.MetricModel.FindListByIds(l.ctx, in.Beid, in.Ids)
	if err != nil {
		return nil, err
	}
	return &metric.MetricListReply{List: metricListToPb(list), Total: int64(len(list))}, nil
}
