package logic

import "errors"

// 业务错误（对齐 quality 服务写法：errors.New 中文语义错误，调用方用 errors.Is 判断）
var (
	// ErrInvalidIdentifier 库名/表名未通过白名单校验（防 SQL 注入）
	ErrInvalidIdentifier = errors.New("表名/库名仅允许字母、数字、下划线")
	// ErrTaskNameRequired 归档任务名缺失
	ErrTaskNameRequired = errors.New("任务名不能为空")
	// ErrDatasourceRequired 数据源缺失
	ErrDatasourceRequired = errors.New("datasource_id 不能为空")
	// ErrTableNameRequired 源表缺失
	ErrTableNameRequired = errors.New("table_name 不能为空")
	// ErrArchiveTableRequired 归档表缺失
	ErrArchiveTableRequired = errors.New("archive_table 不能为空")
	// ErrSameTable 源表与归档表相同（insert select 自循环会造成数据重复/损坏）
	ErrSameTable = errors.New("archive_table 不能与 table_name 相同")
	// ErrInvalidCondition 归档条件非法（须为纯 where 片段且不在黑名单）
	ErrInvalidCondition = errors.New("condition_sql 非法：须以 where 开头，且禁止分号/注释/drop 等关键字")
	// ErrConditionTooLong 归档条件超长（列宽 varchar(1024)，按字符计）
	ErrConditionTooLong = errors.New("condition_sql 过长：不能超过 1024 字符")
	// ErrDataSourceDsn 目标数据源连接串不可用
	ErrDataSourceDsn = errors.New("目标数据源连接串不可用")
)
