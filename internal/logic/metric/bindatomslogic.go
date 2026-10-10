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

type BindAtomsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 派生指标绑定原子指标
func NewBindAtomsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindAtomsLogic {
	return &BindAtomsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 派生指标绑定原子指标: 原子指标 id 切片拷贝一份再转发, 避免与网关请求共享底层数组;
// pb BindAtoms 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *BindAtomsLogic) BindAtoms(req *types.BindAtomsReq) error {
	atomMetricIds := make([]int64, 0, len(req.AtomMetricIds))
	atomMetricIds = append(atomMetricIds, req.AtomMetricIds...)
	_, err := l.svcCtx.MetricRpc.BindAtoms(l.ctx, &metricsvc.BindAtomReq{
		DeriveMetricId: req.DeriveMetricId,
		AtomMetricIds:  atomMetricIds,
		Beid:           req.Beid,
	})
	return err
}
