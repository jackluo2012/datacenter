package logic

import "errors"

// 业务错误（对齐 quality 服务写法：errors.New 中文语义错误，调用方用 errors.Is 判断）
var (
	// ErrIdRequired 通用 id 缺失
	ErrIdRequired = errors.New("id 不能为空")
	// ErrWordRootNameRequired 词根名缺失
	ErrWordRootNameRequired = errors.New("词根名不能为空")
	// ErrWordRootTypeInvalid 词根类型不在白名单（common/attribute/dimension/indicator）
	ErrWordRootTypeInvalid = errors.New("词根类型非法：仅支持 common/attribute/dimension/indicator")
	// ErrWordRootDuplicate 词根名已存在（同 beid 下）
	ErrWordRootDuplicate = errors.New("词根名已存在")
	// ErrWordRootNotFound 词根不存在（含跨租户不可见）
	ErrWordRootNotFound = errors.New("词根不存在")
	// ErrDictCodeRequired 字典编码缺失
	ErrDictCodeRequired = errors.New("字典编码不能为空")
	// ErrDictNameRequired 字典名缺失
	ErrDictNameRequired = errors.New("字典名不能为空")
	// ErrDictItemInvalid 字典项不完整：item_code/item_value 必填且同字典内编码不重复
	ErrDictItemInvalid = errors.New("字典项不完整：item_code/item_value 必填且编码不重复")
	// ErrDictNotFound 字典不存在（含跨租户不可见）
	ErrDictNotFound = errors.New("字典不存在")
	// ErrElementCodeRequired 数据元编码缺失
	ErrElementCodeRequired = errors.New("数据元编码不能为空")
	// ErrElementNameRequired 数据元名缺失
	ErrElementNameRequired = errors.New("数据元名不能为空")
	// ErrDataTypeInvalid data_type 不在白名单（varchar/int/bigint/decimal/date/datetime/text）
	ErrDataTypeInvalid = errors.New("data_type 非法：仅支持 varchar/int/bigint/decimal/date/datetime/text")
	// ErrDataTypeLengthInvalid 长度/精度不合法（varchar 需 data_length>0；decimal 需 data_length>0）
	ErrDataTypeLengthInvalid = errors.New("长度/精度不合法：varchar 与 decimal 的 data_length 必须大于 0")
	// ErrElementCodeDuplicate 数据元编码已存在（同 beid 下）
	ErrElementCodeDuplicate = errors.New("数据元编码已存在")
	// ErrElementNotFound 数据元不存在（含跨租户不可见）
	ErrElementNotFound = errors.New("数据元不存在")
	// ErrDictRefNotFound 数据元值域引用的字典不存在
	ErrDictRefNotFound = errors.New("值域引用的字典不存在")
	// ErrModelCodeRequired 标准模型编码缺失
	ErrModelCodeRequired = errors.New("标准模型编码不能为空")
	// ErrModelNameRequired 标准模型名缺失
	ErrModelNameRequired = errors.New("标准模型名不能为空")
	// ErrModelCodeDuplicate 标准模型编码已存在（同 beid 下）
	ErrModelCodeDuplicate = errors.New("标准模型编码已存在")
	// ErrModelNotFound 标准模型不存在（含跨租户不可见）
	ErrModelNotFound = errors.New("标准模型不存在")
	// ErrElementsJsonInvalid elements_json 非法（非 JSON 数组/元素缺 element_code/编码重复）
	ErrElementsJsonInvalid = errors.New("elements_json 非法：需为 JSON 数组且 element_code 非空不重复")
	// ErrElementMissing elements_json 引用的数据元在库中不存在
	ErrElementMissing = errors.New("elements_json 引用的数据元不存在")
	// ErrModelNotPublished 标准模型未发布（核对仅针对已发布模型）
	ErrModelNotPublished = errors.New("标准模型未发布，请先发布后再核对")
	// ErrDatasourceRequired 数据源缺失
	ErrDatasourceRequired = errors.New("datasource_id 不能为空")
	// ErrTableNameRequired 目标表名缺失
	ErrTableNameRequired = errors.New("table_name 不能为空")
	// ErrInvalidIdentifier 表名/库名未通过白名单校验（防 SQL 注入）
	ErrInvalidIdentifier = errors.New("表名/库名仅允许字母、数字、下划线")
	// ErrTargetTableMissing 目标表在实际库中不存在（无列可核对）
	ErrTargetTableMissing = errors.New("目标表在实际库中不存在")
	// ErrDataSourceDsn 目标数据源连接串不可用
	ErrDataSourceDsn = errors.New("目标数据源连接串不可用")
)
