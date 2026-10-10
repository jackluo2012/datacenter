package logic

import (
	"context"
	"errors"

	"datacenter/metric/model"
	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type BindAtomsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindAtomsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindAtomsLogic {
	return &BindAtomsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 派生指标绑定原子指标（幂等覆盖）：校验派生指标存在且 type=derive，
// 每个原子指标 id 存在且 type=atom，事务内先删旧引用再批量插入
func (l *BindAtomsLogic) BindAtoms(in *metric.BindAtomReq) (*metric.IdReq, error) {
	if in.DeriveMetricId <= 0 {
		return nil, ErrMetricNotFound
	}
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if len(in.AtomMetricIds) == 0 {
		return nil, ErrAtomIdsRequired
	}

	// 1. 派生指标存在且 type=derive
	derive, err := l.svcCtx.MetricModel.FindOne(l.ctx, in.DeriveMetricId, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrMetricNotFound
		}
		return nil, err
	}
	if derive.MetricType != model.MetricTypeDerive {
		return nil, ErrNotDeriveMetric
	}

	// 2. 原子指标 id 去重后批量校验：存在且 type=atom
	atomIds := model.UniqInt64s(in.AtomMetricIds)
	atoms, err := l.svcCtx.MetricModel.FindListByIds(l.ctx, in.Beid, atomIds)
	if err != nil {
		return nil, err
	}
	if len(atoms) != len(atomIds) {
		return nil, ErrAtomMetricMissing
	}
	for _, a := range atoms {
		if a.MetricType != model.MetricTypeAtom {
			return nil, ErrNotAtomMetric
		}
	}

	// 3. 单事务覆盖绑定：先删旧引用再批量插入（幂等）
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		return l.svcCtx.MetricRelModel.BindAll(txCtx, session, in.DeriveMetricId, in.Beid, atomIds)
	})
	if err != nil {
		return nil, err
	}

	l.Infof("[BindAtoms] derive_id=%d atoms=%d beid=%d", in.DeriveMetricId, len(atomIds), in.Beid)
	return &metric.IdReq{Id: in.DeriveMetricId, Beid: in.Beid}, nil
}
