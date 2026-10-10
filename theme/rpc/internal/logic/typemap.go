package logic

// typemap.go 数仓模型字段类型 → 物理列类型映射（纯函数，供建表 DDL 生成用）

// 字段类型白名单（fields_json 每项 type 的合法取值）
const (
	FieldTypeString   = "string"
	FieldTypeInt      = "int"
	FieldTypeBigint   = "bigint"
	FieldTypeDecimal  = "decimal"
	FieldTypeBool     = "bool"
	FieldTypeDate     = "date"
	FieldTypeDatetime = "datetime"
)

// fieldTypes 白名单集合
var fieldTypes = map[string]struct{}{
	FieldTypeString:   {},
	FieldTypeInt:      {},
	FieldTypeBigint:   {},
	FieldTypeDecimal:  {},
	FieldTypeBool:     {},
	FieldTypeDate:     {},
	FieldTypeDatetime: {},
}

// sqlTypes 字段类型 → 物理列类型映射表（建表按此落列：
// string→varchar(255)/int→int/bigint→bigint/decimal→decimal(18,4)/
// bool→tinyint(1)/date→date/datetime→datetime(3)）
var sqlTypes = map[string]string{
	FieldTypeString:   "varchar(255)",
	FieldTypeInt:      "int",
	FieldTypeBigint:   "bigint",
	FieldTypeDecimal:  "decimal(18,4)",
	FieldTypeBool:     "tinyint(1)",
	FieldTypeDate:     "date",
	FieldTypeDatetime: "datetime(3)",
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
