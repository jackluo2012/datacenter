package logic

import "errors"

// 业务错误：调用方可用 errors.Is 判断语义（对齐 votes 服务的业务错误风格）
var (
	// ErrAppNotFound 应用不存在（含跨租户按不存在处理）
	ErrAppNotFound = errors.New("应用不存在")
	// ErrSignInvalid 签名校验失败（签名不匹配 / ts 缺失、非法或超出时间窗口）
	ErrSignInvalid = errors.New("签名校验失败")
	// ErrIpDenied 来源 IP 不在应用白名单内
	ErrIpDenied = errors.New("来源 IP 不在白名单内，拒绝访问")
	// ErrRateLimited 触发应用级限流
	ErrRateLimited = errors.New("请求超过应用限流阈值，请稍后重试")
	// ErrNotPublished API 未发布（draft/offline），拒绝执行
	ErrNotPublished = errors.New("API 未发布，拒绝执行")
	// ErrNotGranted 应用未获得该 API 授权，或授权已过期
	ErrNotGranted = errors.New("应用未获得该 API 授权或授权已过期")
	// ErrParamInvalid 参数校验失败（params_json 定义非法 / 业务参数缺必填、多余、类型错误等）
	ErrParamInvalid = errors.New("参数校验失败")
	// ErrPathDuplicate API 路径在同租户下已存在
	ErrPathDuplicate = errors.New("API 路径已存在")
	// ErrApiNotFound API 不存在
	ErrApiNotFound = errors.New("API 不存在")
	// ErrSqlEmpty SQL 模板为空（发布前置校验）
	ErrSqlEmpty = errors.New("SQL 模板为空")
	// ErrDatasource 获取目标数据源/DSN 失败
	ErrDatasource = errors.New("获取数据源连接失败")
	// ErrSqlExec 目标库 SQL 执行失败
	ErrSqlExec = errors.New("SQL 执行失败")
	// ErrSecretCrypto 应用密钥 AES-GCM 加解密失败
	ErrSecretCrypto = errors.New("应用密钥加解密失败")
)
