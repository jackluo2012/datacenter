package logic

import "errors"

// 业务错误（对齐 metadata 服务写法：errors.New 中文语义错误）
var (
	// ErrBeidRequired 多租户 beid 缺失
	ErrBeidRequired = errors.New("beid 不能为空")
	// ErrNameEnRequired 指标英文名必填
	ErrNameEnRequired = errors.New("指标英文名 name_en 不能为空")
	// ErrNameCnRequired 指标中文名必填
	ErrNameCnRequired = errors.New("指标中文名 name_cn 不能为空")
	// ErrMetricTypeRequired 指标类型必填
	ErrMetricTypeRequired = errors.New("指标类型 metric_type 不能为空")
	// ErrInvalidMetricType 指标类型白名单外（仅支持 atom/derive）
	ErrInvalidMetricType = errors.New("指标类型 metric_type 仅支持 atom/derive")
	// ErrBizFieldRequired 原子指标业务字段必填
	ErrBizFieldRequired = errors.New("原子指标必须指定业务字段 biz_field")
	// ErrAggTypeRequired 原子指标聚合方式必填
	ErrAggTypeRequired = errors.New("原子指标必须指定聚合方式 agg_type")
	// ErrInvalidAggType 聚合方式白名单外
	ErrInvalidAggType = errors.New("聚合方式 agg_type 仅支持 sum/count/avg/max/min/distinct")
	// ErrExprRequired 派生指标表达式必填
	ErrExprRequired = errors.New("派生指标必须指定表达式 expr")
	// ErrDuplicateMetric 指标英文名重复（同 beid 下 name_en 唯一）
	ErrDuplicateMetric = errors.New("指标英文名 name_en 已存在")
	// ErrMetricNotFound 指标不存在
	ErrMetricNotFound = errors.New("指标不存在")
	// ErrNotDeriveMetric 目标指标不是派生指标
	ErrNotDeriveMetric = errors.New("该指标不是派生指标")
	// ErrNotAtomMetric 存在非原子指标，无法绑定为原子引用
	ErrNotAtomMetric = errors.New("绑定的指标中存在非原子指标")
	// ErrAtomIdsRequired 绑定的原子指标列表为空
	ErrAtomIdsRequired = errors.New("绑定的原子指标 id 不能为空")
	// ErrAtomMetricMissing 绑定的原子指标不存在（或不属于当前 beid）
	ErrAtomMetricMissing = errors.New("绑定的原子指标不存在")
	// ErrAtomReferenced 原子指标被派生指标引用，禁止删除
	ErrAtomReferenced = errors.New("原子指标已被派生指标引用，禁止删除")
)
