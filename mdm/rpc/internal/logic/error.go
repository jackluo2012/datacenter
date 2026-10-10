package logic

import "errors"

// 业务错误（对齐 tag/metadata 服务写法：errors.New 中文语义错误）
var (
	// ErrBeidRequired 租户 beid 缺失
	ErrBeidRequired = errors.New("租户 beid 不能为空")
	// ErrInvalidParams 通用入参缺失
	ErrInvalidParams = errors.New("参数不完整")
	// ErrModelNotFound 主数据模型不存在
	ErrModelNotFound = errors.New("主数据模型不存在")
	// ErrDupModelCode 模型编码重复（同租户下 model_code 唯一）
	ErrDupModelCode = errors.New("模型编码已存在")
	// ErrInvalidModelCode 模型编码不合法（须 ^[a-z][a-z0-9_]*$，用于物理表名拼接）
	ErrInvalidModelCode = errors.New("模型编码不合法（须以小写字母开头，仅含小写字母/数字/下划线）")
	// ErrInvalidTableName 物理表名不合法（须 ^[a-z][a-z0-9_]*$，防注入——会真实建表）
	ErrInvalidTableName = errors.New("表名不合法（须以小写字母开头，仅含小写字母/数字/下划线）")
	// ErrInvalidFieldsJson 模型字段定义校验未通过
	ErrInvalidFieldsJson = errors.New("fields_json 校验未通过")
	// ErrInvalidCodeStrategy 编码策略不在白名单（仅 segment/auto）
	ErrInvalidCodeStrategy = errors.New("code_strategy 仅支持 segment/auto")
	// ErrModelInUse 模型下存在主数据记录，禁止删除
	ErrModelInUse = errors.New("模型下存在主数据记录，禁止删除")
	// ErrRecordNotFound 主数据记录不存在
	ErrRecordNotFound = errors.New("主数据记录不存在")
	// ErrDupBizCode 业务编码重复（模型内 biz_code 唯一）
	ErrDupBizCode = errors.New("业务编码已存在")
	// ErrInvalidDataJson 记录数据校验未通过
	ErrInvalidDataJson = errors.New("data_json 校验未通过")
	// ErrRecordArchived 记录已归档，禁止操作
	ErrRecordArchived = errors.New("记录已归档，禁止发布")
	// ErrRecordNotPublished 记录未发布，不能分发
	ErrRecordNotPublished = errors.New("记录未发布，不能分发")
	// ErrInvalidIdentifier 分发目标库/表名不合法（防注入）
	ErrInvalidIdentifier = errors.New("分发目标库/表名不合法")
	// ErrDatasource 数据源 rpc 调用失败或未返回 DSN
	ErrDatasource = errors.New("获取目标数据源失败")
	// ErrDistributeExec 分发执行失败（建表/写入目标库）
	ErrDistributeExec = errors.New("分发执行失败")
	// ErrSegmentStepInvalid 码段步长非法（<=0 无法推进流水）
	ErrSegmentStepInvalid = errors.New("码段步长必须大于 0")
)
