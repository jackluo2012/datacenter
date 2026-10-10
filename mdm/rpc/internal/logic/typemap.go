package logic

// typemap.go 模型字段类型 → 目标库 SQL 列类型映射（纯函数，供自动建表用）

// 字段类型白名单（fields_json 每项 type 的合法取值）
const (
	FieldTypeString  = "string"
	FieldTypeInt     = "int"
	FieldTypeDecimal = "decimal"
	FieldTypeBool    = "bool"
	FieldTypeDate    = "date"
)

// fieldTypes 白名单集合
var fieldTypes = map[string]struct{}{
	FieldTypeString:  {},
	FieldTypeInt:     {},
	FieldTypeDecimal: {},
	FieldTypeBool:    {},
	FieldTypeDate:    {},
}

// sqlTypes 字段类型 → 物理列类型映射表（自动建表按此落列）
var sqlTypes = map[string]string{
	FieldTypeString:  "varchar(255)",
	FieldTypeInt:     "bigint",
	FieldTypeDecimal: "decimal(18,4)",
	FieldTypeBool:    "tinyint(1)",
	FieldTypeDate:    "date",
}

// ValidFieldType 字段类型是否在白名单
func ValidFieldType(t string) bool {
	_, ok := fieldTypes[t]
	return ok
}

// SqlTypeOf 字段类型 → SQL 列类型（未知类型返回空串，调用方先经 ValidFieldType 把关）
func SqlTypeOf(t string) string {
	return sqlTypes[t]
}
