package logic

import (
	"encoding/json"
	"strings"
)

// 数据元数据类型白名单（dc_std_element.data_type）
const (
	DataTypeVarchar  = "varchar"
	DataTypeInt      = "int"
	DataTypeBigint   = "bigint"
	DataTypeDecimal  = "decimal"
	DataTypeDate     = "date"
	DataTypeDatetime = "datetime"
	DataTypeText     = "text"
)

// dataTypeSet data_type 白名单集合
var dataTypeSet = map[string]struct{}{
	DataTypeVarchar:  {},
	DataTypeInt:      {},
	DataTypeBigint:   {},
	DataTypeDecimal:  {},
	DataTypeDate:     {},
	DataTypeDatetime: {},
	DataTypeText:     {},
}

// 词根类型白名单（dc_std_word_root.type）
const (
	WordRootTypeCommon    = "common"
	WordRootTypeAttribute = "attribute"
	WordRootTypeDimension = "dimension"
	WordRootTypeIndicator = "indicator"
)

// wordRootTypeSet 词根类型白名单集合
var wordRootTypeSet = map[string]struct{}{
	WordRootTypeCommon:    {},
	WordRootTypeAttribute: {},
	WordRootTypeDimension: {},
	WordRootTypeIndicator: {},
}

// wordRootTypeDefault 词根类型缺省值（未指定时按通用词根落库）
const wordRootTypeDefault = WordRootTypeCommon

// IsValidDataType 数据元 data_type 是否在白名单内
func IsValidDataType(dataType string) bool {
	_, ok := dataTypeSet[dataType]
	return ok
}

// IsValidWordRootType 词根类型是否在白名单内
func IsValidWordRootType(wordRootType string) bool {
	_, ok := wordRootTypeSet[wordRootType]
	return ok
}

// ModelElement 标准模型组成元素（elements_json 数组元素，含引用与约束）
type ModelElement struct {
	ElementCode string `json:"element_code"` //引用 dc_std_element.element_code
	Required    bool   `json:"required"`     //是否必填标准
	Pk          bool   `json:"pk"`           //是否主键标准
}

// ParseElementsJson 解析并校验 elements_json（纯函数，不查库）：
// 必须为 JSON 数组、每个元素 element_code 非空且数组内不重复。
// 空数组返回空切片不报错（发布时才强制非空）
func ParseElementsJson(elementsJson string) ([]ModelElement, error) {
	s := strings.TrimSpace(elementsJson)
	if s == "" {
		return []ModelElement{}, nil
	}
	var elements []ModelElement
	if err := json.Unmarshal([]byte(s), &elements); err != nil {
		return nil, ErrElementsJsonInvalid
	}
	seen := make(map[string]struct{}, len(elements))
	for _, e := range elements {
		if strings.TrimSpace(e.ElementCode) == "" {
			return nil, ErrElementsJsonInvalid
		}
		if _, dup := seen[e.ElementCode]; dup {
			return nil, ErrElementsJsonInvalid
		}
		seen[e.ElementCode] = struct{}{}
	}
	return elements, nil
}

// ElementCodes 提取元素编码列表（保持原顺序，供批量存在性校验）
func ElementCodes(elements []ModelElement) []string {
	codes := make([]string, 0, len(elements))
	for _, e := range elements {
		codes = append(codes, e.ElementCode)
	}
	return codes
}
