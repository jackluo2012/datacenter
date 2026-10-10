package logic

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// validate.go 纯函数校验：fields_json 模型字段定义、data_json 记录数据、表名/标识符规范化

var (
	//物理表名/模型编码：小写字母开头，仅小写字母/数字/下划线（表名会真实建表，严格防注入）
	identifierRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	//字段名/分发目标库表名：字母开头，仅字母/数字/下划线（合法未加反引号标识符）
	fieldNameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	//日期格式：记录 date 类型字段取值须为 YYYY-MM-DD
	dateLayout = "2006-01-02"
)

// targetTableReservedCols 自动建表的内置列，字段名与之冲突会导致建表重复列，须拒绝
var targetTableReservedCols = map[string]struct{}{
	"id":         {},
	"biz_code":   {},
	"created_at": {},
}

// 编码策略白名单（dc_mdm_model.code_strategy）
const (
	CodeStrategySegment = "segment" //号段模式
	CodeStrategyAuto    = "auto"    //自增
)

// 记录状态（dc_mdm_record.status）
const (
	RecordStatusDraft     = "draft"     //草稿
	RecordStatusPublished = "published" //已发布
	RecordStatusArchived  = "archived"  //已归档
)

// bizCodeMaxLen 业务编码最大长度（dc_mdm_record.biz_code varchar(128)）
const bizCodeMaxLen = 128

// ValidCodeStrategy 编码策略是否在白名单
func ValidCodeStrategy(s string) bool {
	return s == CodeStrategySegment || s == CodeStrategyAuto
}

// trimSpace 去首尾空白
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// upper 去空白后转大写（码段默认前缀用）
func upper(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// ParseFieldsJson 解析已存储的模型字段定义（保存时已校验，此处防御性解析）
func ParseFieldsJson(fieldsJson string) ([]FieldDef, error) {
	var fields []FieldDef
	if err := json.Unmarshal([]byte(fieldsJson), &fields); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFieldsJson, err)
	}
	return fields, nil
}

// marshalFields 序列化字段定义（入参原样合法，统一收敛为规范 JSON 存储）
func marshalFields(fields []FieldDef) (string, error) {
	b, err := json.Marshal(fields)
	if err != nil {
		return "", fmt.Errorf("%w: 序列化失败: %v", ErrInvalidFieldsJson, err)
	}
	return string(b), nil
}

type (
	// FieldDef 模型字段定义（fields_json 数组元素）。
	// name/type 必填；length/required/comment 为可选附加项（对齐 schema 注释约定）
	FieldDef struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Length   int    `json:"length,omitempty"`
		Required bool   `json:"required,omitempty"`
		Comment  string `json:"comment,omitempty"`
	}
)

// ValidIdentifier 物理表名/模型编码是否合法（小写字母开头，仅小写字母/数字/下划线）
func ValidIdentifier(s string) bool {
	return identifierRe.MatchString(s)
}

// ValidFieldName 字段名/分发目标标识符是否合法（字母开头，仅字母/数字/下划线）
func ValidFieldName(s string) bool {
	return fieldNameRe.MatchString(s)
}

// NormalizeTableName 表名规范化（纯函数）：tableName 为空时按 `dc_mdm_data_<model_code>` 自动生成；
// 显式表名与自动生成结果均须通过小写标识符校验（物理表会真实建表，防注入），超 128 字符拒绝
func NormalizeTableName(modelCode, tableName string) (string, error) {
	if !ValidIdentifier(modelCode) {
		return "", ErrInvalidModelCode
	}
	tn := strings.TrimSpace(tableName)
	if tn == "" {
		tn = "dc_mdm_data_" + modelCode
	}
	if !ValidIdentifier(tn) || len(tn) > 128 {
		return "", ErrInvalidTableName
	}
	return tn, nil
}

// ValidateFieldsJson 校验模型字段定义 JSON：须为 JSON 数组，每项为对象且含非空合法的 name
// 与白名单内的 type；name 不与内置列冲突、不重复。返回解析后的字段列表（附加键 length/
// required/comment 合法保留）
func ValidateFieldsJson(fieldsJson string) ([]FieldDef, error) {
	var fields []FieldDef
	if err := json.Unmarshal([]byte(fieldsJson), &fields); err != nil {
		return nil, fmt.Errorf("%w: 须为 JSON 数组: %v", ErrInvalidFieldsJson, err)
	}
	seen := make(map[string]struct{}, len(fields))
	for i, f := range fields {
		if strings.TrimSpace(f.Name) == "" {
			return nil, fmt.Errorf("%w: 第 %d 个字段缺少 name", ErrInvalidFieldsJson, i+1)
		}
		if !ValidFieldName(f.Name) {
			return nil, fmt.Errorf("%w: 字段 %s 名不合法（须以字母开头，仅含字母/数字/下划线）", ErrInvalidFieldsJson, f.Name)
		}
		if _, reserved := targetTableReservedCols[strings.ToLower(f.Name)]; reserved {
			return nil, fmt.Errorf("%w: 字段名 %s 与物理表内置列冲突", ErrInvalidFieldsJson, f.Name)
		}
		if _, dup := seen[f.Name]; dup {
			return nil, fmt.Errorf("%w: 字段 %s 重复定义", ErrInvalidFieldsJson, f.Name)
		}
		seen[f.Name] = struct{}{}
		if !ValidFieldType(f.Type) {
			return nil, fmt.Errorf("%w: 字段 %s 类型 %s 不在白名单（string/int/decimal/bool/date）",
				ErrInvalidFieldsJson, f.Name, f.Type)
		}
	}
	return fields, nil
}

// ValidateDataJson 按模型字段定义校验并收敛记录数据（纯函数）：
//   - data_json 须为 JSON 对象；出现模型未定义的多余字段直接拒绝；
//   - required 字段须存在且非 null/非空串；
//   - 类型收敛：string 收原样；int 收 JSON 数值（须整型）或数字字符串（strconv）；
//     decimal 收数值或数字字符串；bool 收布尔或 "true"/"false"/"1"/"0"；date 须 YYYY-MM-DD；
//   - 返回按模型字段顺序规范化后的 data_json（多余字段不入，缺失的非必填字段不出现在 JSON 中，
//     落目标表时补 NULL）
func ValidateDataJson(fields []FieldDef, dataJson string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(dataJson), &raw); err != nil {
		return "", fmt.Errorf("%w: 须为 JSON 对象: %v", ErrInvalidDataJson, err)
	}

	//多余字段拒绝（以模型字段定义为准）
	fieldSet := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		fieldSet[f.Name] = struct{}{}
	}
	for k := range raw {
		if _, ok := fieldSet[k]; !ok {
			return "", fmt.Errorf("%w: 多余字段 %s（模型未定义）", ErrInvalidDataJson, k)
		}
	}

	//逐字段校验收敛，并按模型字段顺序产出规范化值
	normalized := make(map[string]any, len(fields))
	for _, f := range fields {
		v, ok := raw[f.Name]
		if !ok || v == nil {
			if f.Required {
				return "", fmt.Errorf("%w: 必填字段 %s 缺失", ErrInvalidDataJson, f.Name)
			}
			continue //非必填缺失/null：不入 JSON，落目标表时为 NULL
		}
		//必填字符串字段空串/纯空白视为缺失
		if f.Required {
			if s, isStr := v.(string); isStr && strings.TrimSpace(s) == "" {
				return "", fmt.Errorf("%w: 必填字段 %s 缺失", ErrInvalidDataJson, f.Name)
			}
		}
		cv, err := coerceValue(f, v)
		if err != nil {
			return "", err
		}
		normalized[f.Name] = cv
	}

	out, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("%w: 规范化序列化失败: %v", ErrInvalidDataJson, err)
	}
	return string(out), nil
}

// coerceValue 单字段类型收敛
func coerceValue(f FieldDef, v any) (any, error) {
	switch f.Type {
	case FieldTypeString:
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%w: 字段 %s 须为字符串", ErrInvalidDataJson, f.Name)
		}
		return s, nil
	case FieldTypeInt:
		return coerceInt(f.Name, v)
	case FieldTypeDecimal:
		return coerceDecimal(f.Name, v)
	case FieldTypeBool:
		return coerceBool(f.Name, v)
	case FieldTypeDate:
		return coerceDate(f.Name, v)
	default:
		return nil, fmt.Errorf("%w: 字段 %s 类型 %s 不在白名单", ErrInvalidDataJson, f.Name, f.Type)
	}
}

// coerceInt int 收敛：JSON 整型数值或数字字符串（strconv）
func coerceInt(name string, v any) (any, error) {
	switch n := v.(type) {
	case float64:
		if n != math.Trunc(n) || math.IsInf(n, 0) || math.IsNaN(n) {
			return nil, fmt.Errorf("%w: 字段 %s 须为整型", ErrInvalidDataJson, name)
		}
		return int64(n), nil
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(n), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: 字段 %s 须为整型: %q", ErrInvalidDataJson, name, n)
		}
		return i, nil
	default:
		return nil, fmt.Errorf("%w: 字段 %s 须为整型", ErrInvalidDataJson, name)
	}
}

// coerceDecimal decimal 收敛：JSON 数值或数字字符串（strconv），统一为 float64
func coerceDecimal(name string, v any) (any, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		if err != nil {
			return nil, fmt.Errorf("%w: 字段 %s 须为数值: %q", ErrInvalidDataJson, name, n)
		}
		return f, nil
	default:
		return nil, fmt.Errorf("%w: 字段 %s 须为数值", ErrInvalidDataJson, name)
	}
}

// coerceBool bool 收敛：JSON 布尔或 "true"/"false"/"1"/"0" 字符串（strconv.ParseBool）
func coerceBool(name string, v any) (any, error) {
	switch b := v.(type) {
	case bool:
		return b, nil
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(b))
		if err != nil {
			return nil, fmt.Errorf("%w: 字段 %s 须为布尔: %q", ErrInvalidDataJson, name, b)
		}
		return parsed, nil
	default:
		return nil, fmt.Errorf("%w: 字段 %s 须为布尔", ErrInvalidDataJson, name)
	}
}

// coerceDate date 收敛：字符串且格式为 YYYY-MM-DD（time.Parse 严格校验）
func coerceDate(name string, v any) (any, error) {
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("%w: 字段 %s 须为 YYYY-MM-DD 日期字符串", ErrInvalidDataJson, name)
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return nil, fmt.Errorf("%w: 字段 %s 日期格式须为 YYYY-MM-DD: %q", ErrInvalidDataJson, name, s)
	}
	return t.Format(dateLayout), nil
}
