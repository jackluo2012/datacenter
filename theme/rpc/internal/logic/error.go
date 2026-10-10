package logic

import "errors"

// 业务错误（对齐 mdm/tag 服务写法：errors.New 中文语义错误）
var (
	// ErrBeidRequired 租户 beid 缺失
	ErrBeidRequired = errors.New("租户 beid 不能为空")
	// ErrInvalidParams 通用入参缺失
	ErrInvalidParams = errors.New("参数不完整")
	// ErrInvalidDomainCode 域编码不合法（须小写字母开头，仅小写字母/数字/下划线）
	ErrInvalidDomainCode = errors.New("域编码不合法（须以小写字母开头，仅含小写字母/数字/下划线）")
	// ErrDupDomainCode 域编码重复（同租户下 code 唯一）
	ErrDupDomainCode = errors.New("域编码已存在")
	// ErrDomainNotFound 主题域不存在
	ErrDomainNotFound = errors.New("主题域不存在")
	// ErrDomainInUse 域下存在主题，禁止删除
	ErrDomainInUse = errors.New("域下存在主题，禁止删除")
	// ErrInvalidThemeCode 主题编码不合法（须小写字母开头，仅小写字母/数字/下划线）
	ErrInvalidThemeCode = errors.New("主题编码不合法（须以小写字母开头，仅含小写字母/数字/下划线）")
	// ErrDupThemeCode 主题编码重复（同租户下 code 唯一）
	ErrDupThemeCode = errors.New("主题编码已存在")
	// ErrThemeNotFound 主题不存在
	ErrThemeNotFound = errors.New("主题不存在")
	// ErrThemeInUse 主题下存在模型，禁止删除
	ErrThemeInUse = errors.New("主题下存在模型，禁止删除")
	// ErrInvalidLayer 数仓分层不在白名单（仅 ods/dwd/dws/ads/dim）
	ErrInvalidLayer = errors.New("layer 仅支持 ods/dwd/dws/ads/dim")
	// ErrInvalidTableName 模型表名不合法（须小写字母开头，仅小写字母/数字/下划线——会真实建表）
	ErrInvalidTableName = errors.New("表名不合法（须以小写字母开头，仅含小写字母/数字/下划线）")
	// ErrDupTableName 表名重复（同租户下 table_name 唯一）
	ErrDupTableName = errors.New("模型表名已存在")
	// ErrModelNotFound 数仓模型不存在
	ErrModelNotFound = errors.New("数仓模型不存在")
	// ErrInvalidDbName 目标库名不合法（防注入——库名会进 DDL）
	ErrInvalidDbName = errors.New("目标库名不合法（须以小写字母开头，仅含小写字母/数字/下划线）")
	// ErrInvalidFieldsJson fields_json 校验未通过
	ErrInvalidFieldsJson = errors.New("fields_json 校验未通过")
	// ErrInvalidDimJson dim_json 校验未通过
	ErrInvalidDimJson = errors.New("dim_json 校验未通过")
	// ErrDatasource 数据源 rpc 调用失败或未返回 DSN
	ErrDatasource = errors.New("获取目标数据源失败")
	// ErrDeployExec 建表 DDL 执行失败
	ErrDeployExec = errors.New("模型部署执行失败")
	// ErrMvNotFound 物化视图不存在
	ErrMvNotFound = errors.New("物化视图不存在")
	// ErrInvalidViewName 视图名不合法（须小写字母开头，仅小写字母/数字/下划线——会真实建视图）
	ErrInvalidViewName = errors.New("视图名不合法（须以小写字母开头，仅含小写字母/数字/下划线）")
	// ErrInvalidSelectSql select_sql 校验未通过（须 SELECT 开头且不含分号）
	ErrInvalidSelectSql = errors.New("select_sql 非法（须以 SELECT 开头，且不含分号）")
	// ErrInvalidRefreshMode 刷新模式不在白名单（仅 manual/schedule）
	ErrInvalidRefreshMode = errors.New("refresh_mode 仅支持 manual/schedule")
)
