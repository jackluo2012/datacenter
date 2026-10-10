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

type MetricListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 指标列表
func NewMetricListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MetricListLogic {
	return &MetricListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 指标列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *MetricListLogic) MetricList(req *types.MetricListReq) (resp *types.MetricListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.MetricRpc.ListMetrics(l.ctx, &metricsvc.MetricListReq{
		Beid:       req.Beid,
		MetricType: req.MetricType,
		Keyword:    req.Keyword,
		Page:       page,
		Size:       size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return metricListToTypes(reply), nil
}
