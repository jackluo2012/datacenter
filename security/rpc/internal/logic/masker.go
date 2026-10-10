package logic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// 脱敏算法白名单（dc_sec_mask_rule.algorithm）
const (
	AlgorithmMask    = "mask"    //保留前 prefix 后 suffix 字符，中间以 mask_char 填充
	AlgorithmHash    = "hash"    //sha256 hex 前 16 位
	AlgorithmReplace = "replace" //整值替换为 replace 值
	AlgorithmNull    = "null"    //返回空串
)

// defaultMaskChar mask 算法缺省掩码字符
const defaultMaskChar = "*"

// algorithmSet 算法白名单集合
var algorithmSet = map[string]struct{}{
	AlgorithmMask:    {},
	AlgorithmHash:    {},
	AlgorithmReplace: {},
	AlgorithmNull:    {},
}

// IsValidAlgorithm 算法是否在白名单内
func IsValidAlgorithm(algorithm string) bool {
	_, ok := algorithmSet[algorithm]
	return ok
}

// MaskParams params_json 结构（按 algorithm 取用对应字段；
// Replace/MaskChar 用指针区分"键缺失"与"空串值"——replace 允许整值替换为空串）
type MaskParams struct {
	Prefix   int     `json:"prefix"`    //mask：保留前缀字符数（rune 计）
	Suffix   int     `json:"suffix"`    //mask：保留后缀字符数（rune 计）
	MaskChar *string `json:"mask_char"` //mask：掩码字符，缺省 *，必须单 rune
	Replace  *string `json:"replace"`   //replace：整值替换值（键必填，允许空串）
}

// ParseMaskParams 解析并按算法校验 params_json（纯函数，不查库）：
//
//	mask    → params_json 必填，prefix/suffix 须为非负整数；mask_char 若给出须为单 rune
//	replace → replace 键必须存在（允许空串替换）
//	hash/null → 无参数要求，params_json 可为空
func ParseMaskParams(algorithm, paramsJson string) (*MaskParams, error) {
	p := &MaskParams{}
	s := strings.TrimSpace(paramsJson)
	if s == "" {
		// mask/replace 的参数是语义必需的，缺参直接拒绝
		if algorithm == AlgorithmMask || algorithm == AlgorithmReplace {
			return nil, ErrMaskParamsInvalid
		}
		return p, nil
	}
	if err := json.Unmarshal([]byte(s), p); err != nil {
		return nil, ErrMaskParamsInvalid
	}
	switch algorithm {
	case AlgorithmMask:
		if p.Prefix < 0 || p.Suffix < 0 {
			return nil, ErrMaskParamsInvalid
		}
		if p.MaskChar != nil && len([]rune(*p.MaskChar)) != 1 {
			return nil, ErrMaskParamsInvalid
		}
	case AlgorithmReplace:
		if p.Replace == nil {
			return nil, ErrMaskParamsInvalid
		}
	}
	return p, nil
}

// MaskValue 按算法对样例值脱敏（纯函数，MaskPreview 与结果集脱敏共用）：
//
//	mask    → 保留前 prefix + 后 suffix 个 rune，中间以 mask_char 重复填充（rune 安全，中文不乱码）；
//	          长度不足或前后缀重叠时整值掩码，避免短值原样透出
//	hash    → sha256 hex 前 16 位
//	replace → 整值替换为 replace 值
//	null    → 返回空串
//	未知算法 → 返回空串兜底（fail-closed，白名单校验正常在保存侧拦截）
func MaskValue(algorithm, sample string, p *MaskParams) string {
	switch algorithm {
	case AlgorithmMask:
		if p == nil {
			p = &MaskParams{}
		}
		return maskByRunes(sample, p.Prefix, p.Suffix, maskCharOf(p))
	case AlgorithmHash:
		sum := sha256.Sum256([]byte(sample))
		return hex.EncodeToString(sum[:])[:16]
	case AlgorithmReplace:
		if p == nil || p.Replace == nil {
			return ""
		}
		return *p.Replace
	case AlgorithmNull:
		return ""
	default:
		// 未知算法按全掩码兜底（fail-closed）
		return strings.Repeat(defaultMaskChar, len([]rune(sample)))
	}
}

// maskByRunes mask 算法核心：按 rune 计数保留前后缀，中间填充掩码字符；
// 值长度 ≤ prefix+suffix（含前后缀重叠）时整值掩码，杜绝短值漏脱敏；
// 负数参数钳位为 0（防御直调，正常链路已在 ParseMaskParams 拦截）
func maskByRunes(sample string, prefix, suffix int, maskChar string) string {
	if prefix < 0 {
		prefix = 0
	}
	if suffix < 0 {
		suffix = 0
	}
	runes := []rune(sample)
	n := len(runes)
	if n <= prefix+suffix {
		return strings.Repeat(maskChar, n)
	}
	head := runes[:prefix]
	tail := runes[n-suffix:]
	return string(head) + strings.Repeat(maskChar, n-prefix-suffix) + string(tail)
}

// maskCharOf 掩码字符：未配置或配置为空时用缺省 "*"
func maskCharOf(p *MaskParams) string {
	if p != nil && p.MaskChar != nil && *p.MaskChar != "" {
		return *p.MaskChar
	}
	return defaultMaskChar
}
