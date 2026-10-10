package logic

import (
	"datacenter/metric/model"
	"datacenter/metric/rpc/metric"
)

// metricToPb 指标 model -> pb
func metricToPb(m *model.Metric) *metric.Metric {
	return &metric.Metric{
		Id:            m.Id,
		NameEn:        m.NameEn,
		NameCn:        m.NameCn,
		MetricType:    m.MetricType,
		BizField:      m.BizField,
		AggType:       m.AggType,
		Expr:          m.Expr,
		DimensionJson: m.DimensionJson,
		Comment:       m.Comment,
		Owner:         m.Owner,
		Status:        int32(m.Status),
		Beid:          m.Beid,
	}
}

// metricListToPb 指标列表 model -> pb（空列表返回空切片，避免 JSON 序列化成 null）
func metricListToPb(list []*model.Metric) []*metric.Metric {
	out := make([]*metric.Metric, 0, len(list))
	for _, m := range list {
		out = append(out, metricToPb(m))
	}
	return out
}
