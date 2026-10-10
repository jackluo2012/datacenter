package logic

import "errors"

// 业务错误（对齐 standard 服务写法：errors.New 中文语义错误，调用方用 errors.Is 判断）
var (
	// ErrIdRequired 通用 id 缺失
	ErrIdRequired = errors.New("id 不能为空")
	// ErrLevelCodeInvalid 分级编码未通过白名单（需匹配 ^L[1-9]$）
	ErrLevelCodeInvalid = errors.New("code 非法：需为 L1-L9")
	// ErrLevelNameRequired 分级名缺失
	ErrLevelNameRequired = errors.New("name 不能为空")
	// ErrLevelRefNotFound 密级编码不在分级目录中
	ErrLevelRefNotFound = errors.New("密级编码不在分级目录中")
	// ErrEntityIdRequired 打标实体 id 缺失
	ErrEntityIdRequired = errors.New("entity_id 不能为空")
	// ErrMaskRuleRefNotFound 打标绑定的脱敏规则不存在
	ErrMaskRuleRefNotFound = errors.New("绑定的脱敏规则不存在")
	// ErrRuleCodeRequired 脱敏规则编码缺失
	ErrRuleCodeRequired = errors.New("rule_code 不能为空")
	// ErrRuleNameRequired 脱敏规则名缺失
	ErrRuleNameRequired = errors.New("rule_name 不能为空")
	// ErrAlgorithmInvalid 脱敏算法不在白名单（mask/hash/replace/null）
	ErrAlgorithmInvalid = errors.New("algorithm 非法：仅支持 mask/hash/replace/null")
	// ErrMaskParamsInvalid params_json 与算法要求不匹配
	// （mask 需非负整数 prefix/suffix 且掩码字符为单 rune；replace 需携带 replace 键）
	ErrMaskParamsInvalid = errors.New("params_json 非法：与算法要求不匹配")
	// ErrRuleCodeDuplicate 脱敏规则编码已存在（同 beid 下）
	ErrRuleCodeDuplicate = errors.New("脱敏规则编码已存在")
	// ErrRuleNotFound 脱敏规则不存在（含跨租户不可见）
	ErrRuleNotFound = errors.New("脱敏规则不存在")
	// ErrRuleInUse 脱敏规则已被实体打标引用，禁止删除
	ErrRuleInUse = errors.New("脱敏规则已被实体打标引用，禁止删除")
	// ErrSubjectInvalid 授权主体格式非法（需为 user:id/app:id/role:id）
	ErrSubjectInvalid = errors.New("subject 非法：需为 user:id/app:id/role:id 格式")
	// ErrResourceRequired 授权资源类型缺失
	ErrResourceRequired = errors.New("resource 不能为空")
	// ErrGrantNotFound 授权记录不存在（含跨租户不可见）
	ErrGrantNotFound = errors.New("授权记录不存在")
)
