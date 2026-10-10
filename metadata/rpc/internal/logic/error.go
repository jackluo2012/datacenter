package logic

import "errors"

// 业务错误（对齐 votes 服务写法：errors.New 中文语义错误）
var (
	// ErrInvalidCollectParams 采集入参缺失（db_type/dsn/db_name/beid）
	ErrInvalidCollectParams = errors.New("采集参数不完整：db_type/dsn/db_name/beid 均不能为空")
	// ErrEntityNotFound 实体不存在
	ErrEntityNotFound = errors.New("实体不存在")
	// ErrInvalidLineageParams 血缘起点/终点非法
	ErrInvalidLineageParams = errors.New("血缘起点/终点实体id不能为空")
	// ErrInvalidDirection 血缘方向非法
	ErrInvalidDirection = errors.New("direction 仅支持 upstream/downstream")
	// ErrVersionNotFound 版本不存在
	ErrVersionNotFound = errors.New("版本不存在")
)
