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

// 删除指标：派生指标级联删 rel；原子指标被派生引用时返回业务错误拒绝删除
func (l *DeleteMetricLogic) DeleteMetric(in *metric.IdReq) (*metric.IdReq, error) {
	if in.Id <= 0 {
		return nil, ErrMetricNotFound
	}
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	m, err := l.svcCtx.MetricModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrMetricNotFound
		}
		return nil, err
	}

	// 原子指标：仍被派生指标引用时禁止删除
	if m.MetricType == model.MetricTypeAtom {
		cnt, err := l.svcCtx.MetricRelModel.CountByAtomId(l.ctx, m.Id)
		if err != nil {
			return nil, err
		}
		if cnt > 0 {
			return nil, ErrAtomReferenced
		}
	}

	// 派生指标：单事务先删 rel 再删本体；原子指标直接删本体
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		if m.MetricType == model.MetricTypeDerive {
			if err := l.svcCtx.MetricRelModel.DeleteByDeriveId(txCtx, session, m.Id); err != nil {
				return err
			}
		}
		return l.svcCtx.MetricModel.Delete(txCtx, session, m.Id, in.Beid)
	})
	if err != nil {
		return nil, err
	}

	l.Infof("[DeleteMetric] id=%d type=%s beid=%d", m.Id, m.MetricType, in.Beid)
	return &metric.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
