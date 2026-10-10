package metric

import (
	"datacenter/internal/types"
	"datacenter/metric/rpc/metricsvc"
)

// metricToTypes 将 rpc 的 Metric 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func metricToTypes(m *metricsvc.Metric) *types.MetricInfo {
	if m == nil {
		return &types.MetricInfo{}
	}
	return &types.MetricInfo{
		Id:            m.GetId(),
		NameEn:        m.GetNameEn(),
		NameCn:        m.GetNameCn(),
		MetricType:    m.GetMetricType(),
		BizField:      m.GetBizField(),
		AggType:       m.GetAggType(),
		Expr:          m.GetExpr(),
		DimensionJson: m.GetDimensionJson(),
		Comment:       m.GetComment(),
		Owner:         m.GetOwner(),
		Status:        m.GetStatus(),
	}
}

// metricListToTypes 将 rpc 的指标列表逐字段转换为网关类型, 供指标列表与原子指标查询复用。
func metricListToTypes(reply *metricsvc.MetricListReply) *types.MetricListResp {
	list := make([]types.MetricInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *metricToTypes(item))
	}
	return &types.MetricListResp{
		List:  list,
		Total: reply.GetTotal(),
	}
}
