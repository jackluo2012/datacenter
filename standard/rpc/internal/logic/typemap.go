package logic

import (
	"regexp"
	"strconv"
	"strings"
)

// 差异结果（diff_json 数组元素的 result 字段）
const (
	DiffResultMatch        = "match"         // 列存在且类型匹配
	DiffResultMissing      = "missing"       // 列缺失
	DiffResultTypeMismatch = "type_mismatch" // 列存在但类型不匹配
)

// identifierRe 标识符白名单：表名/库名仅允许字母、数字、下划线，
// 不通过者一律拒绝拼接（杜绝 `users; drop table x` 类注入）
var identifierRe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// ValidIdentifier 标识符（库名/表名）白名单校验，防 SQL 注入
func ValidIdentifier(name string) bool {
	return identifierRe.MatchString(name)
}

// StdColumn 标准期望列（由数据元定义映射而来，纯函数入参，不依赖 model）
type StdColumn struct {
	Name      string //列名（数据元编码）
	DataType  string //标准类型 varchar/int/bigint/decimal/date/datetime/text
	Length    int32  //标准长度（varchar 字符长度；decimal 整数位总精度）
	Precision int32  //标准小数位（decimal 用）
}

// ActualColumn 实际列（information_schema.COLUMNS 采集结果，纯函数入参）
type ActualColumn struct {
	Name         string //COLUMN_NAME
	DataType     string //DATA_TYPE 基础类型（varchar/int/decimal/...，不含 unsigned 等修饰）
	ColumnType   string //COLUMN_TYPE 完整类型（varchar(50)/decimal(10,2)/int unsigned）
	CharLen      int64  //CHARACTER_MAXIMUM_LENGTH（字符类型有值）
	NumPrecision int64  //NUMERIC_PRECISION（数值类型总精度）
	NumScale     int64  //NUMERIC_SCALE（数值类型小数位）
}

// DiffItem 表结构差异项（diff_json 数组元素）
type DiffItem struct {
	Column string `json:"column"` //列名（数据元编码）
	Expect string `json:"expect"` //期望类型串（如 varchar(50)/decimal(10,2)）
	Actual string `json:"actual"` //实际类型串（COLUMN_TYPE，缺列时为空）
	Result string `json:"result"` //match/missing/type_mismatch
}

// ExpectType 生成期望类型串（纯函数）：
//
//	varchar   → varchar(N)
//	decimal   → decimal(P,S)（P=length 总精度，S=precision 小数位）
//	其余类型  → 基础类型名本身（int/bigint/date/datetime/text）
func ExpectType(dataType string, length, precision int32) string {
	switch dataType {
	case DataTypeVarchar:
		if length > 0 {
			return dataType + "(" + strconv.Itoa(int(length)) + ")"
		}
		return dataType
	case DataTypeDecimal:
		if length > 0 {
			return dataType + "(" + strconv.Itoa(int(length)) + "," + strconv.Itoa(int(precision)) + ")"
		}
		return dataType
	default:
		return dataType
	}
}

// TypeMatched 类型映射匹配（纯函数）。规则：
//
//	varchar  → 实际基础类型同为 varchar 且字符长度一致（varchar(50) vs varchar(100) 视为不匹配）
//	decimal  → 实际基础类型同为 decimal，总精度与标准 length 一致（length>0 时比对），
//	           小数位与标准 precision 一致
//	int/bigint/date/datetime/text → 实际基础类型相同即可（可容纳性由标准侧定义保证）
//	其余/未知类型 → 一律不匹配
func TypeMatched(dataType string, length, precision int32, actual ActualColumn) bool {
	switch dataType {
	case DataTypeVarchar:
		if actual.DataType != DataTypeVarchar {
			return false
		}
		return length <= 0 || actual.CharLen == int64(length)
	case DataTypeDecimal:
		if actual.DataType != DataTypeDecimal {
			return false
		}
		if length > 0 && actual.NumPrecision != int64(length) {
			return false
		}
		return actual.NumScale == int64(precision)
	case DataTypeInt, DataTypeBigint, DataTypeDate, DataTypeDatetime, DataTypeText:
		return actual.DataType == dataType
	default:
		return false
	}
}

// actualTypeText 实际类型串：优先 COLUMN_TYPE（含长度/符号修饰），缺失时退回 DATA_TYPE
func actualTypeText(actual ActualColumn) string {
	s := strings.TrimSpace(actual.ColumnType)
	if s == "" {
		return actual.DataType
	}
	return s
}

// BuildStructDiff 表结构比对（纯函数）：按标准列序逐一比对期望与实际，
// 输出差异项列表与整体通过标志（全部 match 才通过）。
// actuals 按列名小写建索引比对（MySQL 列名大小写不敏感）
func BuildStructDiff(stdCols []*StdColumn, actuals []*ActualColumn) ([]*DiffItem, bool) {
	index := make(map[string]*ActualColumn, len(actuals))
	for _, a := range actuals {
		index[strings.ToLower(a.Name)] = a
	}

	pass := true
	items := make([]*DiffItem, 0, len(stdCols))
	for _, std := range stdCols {
		item := &DiffItem{
			Column: std.Name,
			Expect: ExpectType(std.DataType, std.Length, std.Precision),
		}
		actual, ok := index[strings.ToLower(std.Name)]
		switch {
		case !ok:
			item.Result = DiffResultMissing
		case TypeMatched(std.DataType, std.Length, std.Precision, *actual):
			item.Actual = actualTypeText(*actual)
			item.Result = DiffResultMatch
		default:
			item.Actual = actualTypeText(*actual)
			item.Result = DiffResultTypeMismatch
		}
		if item.Result != DiffResultMatch {
			pass = false
		}
		items = append(items, item)
	}
	return items, pass
}
