// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package metric

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/metric/rpc/metricsvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MetricAtomsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 派生指标引用的原子指标
func NewMetricAtomsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MetricAtomsLogic {
	return &MetricAtomsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 派生指标引用的原子指标: 转发 ListAtomsOfDerive, 复用指标列表映射转为网关类型返回
func (l *MetricAtomsLogic) MetricAtoms(req *types.MetricIdReq) (resp *types.MetricListResp, err error) {
	reply, err := l.svcCtx.MetricRpc.ListAtomsOfDerive(l.ctx, &metricsvc.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return metricListToTypes(reply), nil
}
