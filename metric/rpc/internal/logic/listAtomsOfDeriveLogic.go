package logic

import (
	"context"
	"errors"

	"datacenter/metric/model"
	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAtomsOfDeriveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAtomsOfDeriveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAtomsOfDeriveLogic {
	return &ListAtomsOfDeriveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取派生指标引用的原子指标：先确认指标存在且 type=derive，再按 rel 反查原子指标回填
func (l *ListAtomsOfDeriveLogic) ListAtomsOfDerive(in *metric.IdReq) (*metric.MetricListReply, error) {
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
	if m.MetricType != model.MetricTypeDerive {
		return nil, ErrNotDeriveMetric
	}

	atomIds, err := l.svcCtx.MetricRelModel.ListAtoms(l.ctx, m.Id)
	if err != nil {
		return nil, err
	}
	if len(atomIds) == 0 {
		return &metric.MetricListReply{List: []*metric.Metric{}}, nil
	}

	atoms, err := l.svcCtx.MetricModel.FindListByIds(l.ctx, in.Beid, atomIds)
	if err != nil {
		return nil, err
	}
	return &metric.MetricListReply{List: metricListToPb(atoms), Total: int64(len(atoms))}, nil
}
