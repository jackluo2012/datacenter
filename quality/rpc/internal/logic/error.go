package logic

import "errors"

// 业务错误（对齐 metadata 服务写法：errors.New 中文语义错误，调用方用 errors.Is 判断）
var (
	// ErrInvalidRuleType 规则类型不在白名单
	ErrInvalidRuleType = errors.New("rule_type 非法：仅支持 row_count/not_null/unique/value_range/consistency/fluctuation")
	// ErrInvalidRuleParams 规则参数缺失或非法（按 rule_type 校验 target_column/params_json 必填性）
	ErrInvalidRuleParams = errors.New("规则参数不完整或非法")
	// ErrInvalidIdentifier 表名/列名/库名未通过白名单校验（防 SQL 注入）
	ErrInvalidIdentifier = errors.New("表名/列名仅允许字母、数字、下划线")
	// ErrRuleNameRequired 规则名缺失
	ErrRuleNameRequired = errors.New("规则名不能为空")
	// ErrTargetTableRequired 目标表缺失
	ErrTargetTableRequired = errors.New("目标表不能为空")
	// ErrRuleNotFound 规则不存在（含跨租户不可见）
	ErrRuleNotFound = errors.New("规则不存在")
	// ErrJobNameRequired 任务名缺失
	ErrJobNameRequired = errors.New("任务名不能为空")
	// ErrDatasourceRequired 数据源缺失
	ErrDatasourceRequired = errors.New("datasource_id 不能为空")
	// ErrEmptyRuleIds 任务未配置规则
	ErrEmptyRuleIds = errors.New("任务未配置质量规则：rule_ids 不能为空")
	// ErrInvalidRuleIds rule_ids 非法（非逗号分隔的正整数）
	ErrInvalidRuleIds = errors.New("rule_ids 非法：需为逗号分隔的正整数id")
	// ErrDataSourceDsn 目标数据源连接串不可用
	ErrDataSourceDsn = errors.New("目标数据源连接串不可用")
)
