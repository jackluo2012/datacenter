package logic

import "datacenter/metric/model"

// aggTypeWhitelist 聚合方式白名单（对应 dc_metric.agg_type 注释口径）
var aggTypeWhitelist = map[string]struct{}{
	"sum":      {},
	"count":    {},
	"avg":      {},
	"max":      {},
	"min":      {},
	"distinct": {},
}

// IsValidMetricType 指标类型校验（仅支持 atom/derive）
func IsValidMetricType(metricType string) bool {
	return metricType == model.MetricTypeAtom || metricType == model.MetricTypeDerive
}

// IsValidAggType 聚合方式白名单校验
func IsValidAggType(aggType string) bool {
	_, ok := aggTypeWhitelist[aggType]
	return ok
}

// validateMetricSpec 按指标类型校验口径字段（纯函数，创建指标复用，单测覆盖）：
//   - atom：biz_field + agg_type 必填，且 agg_type 必须在白名单内
//   - derive：expr 必填
func validateMetricSpec(metricType, bizField, aggType, expr string) error {
	if metricType == model.MetricTypeAtom {
		if bizField == "" {
			return ErrBizFieldRequired
		}
		if aggType == "" {
			return ErrAggTypeRequired
		}
		if !IsValidAggType(aggType) {
			return ErrInvalidAggType
		}
		return nil
	}
	// derive
	if expr == "" {
		return ErrExprRequired
	}
	return nil
}
