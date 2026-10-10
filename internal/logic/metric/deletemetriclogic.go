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

type DeleteMetricLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除指标
func NewDeleteMetricLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMetricLogic {
	return &DeleteMetricLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除指标: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteMetric 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteMetricLogic) DeleteMetric(req *types.MetricIdReq) error {
	_, err := l.svcCtx.MetricRpc.DeleteMetric(l.ctx, &metricsvc.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
