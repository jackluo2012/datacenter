package logic

import "errors"

// 业务错误（对齐 metric/tag 服务写法：errors.New 中文语义错误）
var (
	// ErrBeidRequired 租户 beid 缺失
	ErrBeidRequired = errors.New("beid 不能为空")
	// ErrNameRequired 脚本名必填
	ErrNameRequired = errors.New("脚本名 name 不能为空")
	// ErrScriptNotFound 脚本不存在（或不属于当前租户）
	ErrScriptNotFound = errors.New("脚本不存在")
	// ErrInvalidSql 脚本 SQL 未通过安全闸（仅允许单条 SELECT）
	ErrInvalidSql = errors.New("脚本 SQL 校验未通过")
	// ErrDatasource 数据源 rpc 调用失败或未返回 DSN
	ErrDatasource = errors.New("获取目标数据源失败")
	// ErrScriptExec 脚本执行失败
	ErrScriptExec = errors.New("脚本执行失败")
)
