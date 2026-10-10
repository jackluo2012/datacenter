package logic

import "errors"

// 业务错误（对齐 metadata/votes 服务写法：errors.New 中文语义错误）
var (
	// ErrBeidRequired 租户 beid 缺失
	ErrBeidRequired = errors.New("租户 beid 不能为空")
	// ErrInvalidParams 通用入参缺失
	ErrInvalidParams = errors.New("参数不完整")
	// ErrInvalidObjectType 对象类型不在白名单（仅 user/member/goods）
	ErrInvalidObjectType = errors.New("object_type 仅支持 user/member/goods")
	// ErrInvalidTagType 标签类型不在白名单（仅 rule/fact/model）
	ErrInvalidTagType = errors.New("tag_type 仅支持 rule/fact/model")
	// ErrObjectNotFound 标签对象不存在
	ErrObjectNotFound = errors.New("标签对象不存在")
	// ErrTagNotFound 标签定义不存在
	ErrTagNotFound = errors.New("标签定义不存在")
	// ErrDupTagKey 同一对象下 tag_key 重复
	ErrDupTagKey = errors.New("同一对象下 tag_key 已存在")
	// ErrInvalidRuleSql 规则 SQL 未通过校验（非 select 开头/缺别名列/含分号等）
	ErrInvalidRuleSql = errors.New("rule_sql 校验未通过")
	// ErrNotRuleTag 非 rule 类型标签无规则 SQL，不能执行打标
	ErrNotRuleTag = errors.New("仅 rule 类型标签支持打标")
	// ErrTagTargetRequired 打标目标缺失（tag_id 与 object_id 至少传一个）
	ErrTagTargetRequired = errors.New("tag_id 与 object_id 至少传一个")
	// ErrObjectTargetMissing 对象未配置数据源/表，无法打标
	ErrObjectTargetMissing = errors.New("标签对象未配置 ds_id/table_name，无法打标")
	// ErrDatasource 数据源 rpc 调用失败或未返回 DSN
	ErrDatasource = errors.New("获取目标数据源失败")
	// ErrRuleExec 规则 SQL 执行失败
	ErrRuleExec = errors.New("规则 SQL 执行失败")
)
