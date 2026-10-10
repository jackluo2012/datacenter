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

type CreateMetricLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建指标
func NewCreateMetricLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMetricLogic {
	return &CreateMetricLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 创建指标: 网关请求逐字段映射为 rpc 请求, 指标类型缺省为 atom(原子指标), 转发后把指标转为网关类型返回
func (l *CreateMetricLogic) CreateMetric(req *types.CreateMetricReq) (resp *types.MetricInfo, err error) {
	metricType := req.MetricType
	if metricType == "" {
		metricType = "atom"
	}
	reply, err := l.svcCtx.MetricRpc.CreateMetric(l.ctx, &metricsvc.CreateMetricReq{
		NameEn:        req.NameEn,
		NameCn:        req.NameCn,
		MetricType:    metricType,
		BizField:      req.BizField,
		AggType:       req.AggType,
		Expr:          req.Expr,
		DimensionJson: req.DimensionJson,
		Comment:       req.Comment,
		Owner:         req.Owner,
		Beid:          req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return metricToTypes(reply.GetMetric()), nil
}
