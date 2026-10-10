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

type MetricInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 指标详情
func NewMetricInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MetricInfoLogic {
	return &MetricInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 指标详情: 网关请求逐字段映射为 rpc 请求, 转发后把指标转为网关类型返回
func (l *MetricInfoLogic) MetricInfo(req *types.MetricIdReq) (resp *types.MetricInfo, err error) {
	reply, err := l.svcCtx.MetricRpc.GetMetric(l.ctx, &metricsvc.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return metricToTypes(reply.GetMetric()), nil
}
