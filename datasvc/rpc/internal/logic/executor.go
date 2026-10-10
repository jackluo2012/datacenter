package logic

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// 参数类型取值（params_json 中 type 仅支持这三种）
const (
	TypeString = "string"
	TypeNumber = "number"
	TypeBool   = "bool"
)

// maskValue 脱敏后的替换值
const maskValue = "****"

// ParamDef API 参数定义（params_json 数组元素）
type ParamDef struct {
	Name     string      `json:"name"`     //参数名，对应 SQL 模板中的 :name 占位符
	Type     string      `json:"type"`     //string/number/bool
	Required bool        `json:"required"` //是否必填
	Default  interface{} `json:"default"`  //缺省值（非必填且未传时补全）
}

// paramNameRe SQL 模板中的占位符形如 :name（字母/下划线开头）
var paramNameRe = regexp.MustCompile(`:[A-Za-z_][A-Za-z0-9_]*`)

// ParseParamDefs 解析并校验 params_json 参数定义（空串视为无参数）
func ParseParamDefs(paramsJson string) ([]ParamDef, error) {
	paramsJson = strings.TrimSpace(paramsJson)
	if paramsJson == "" {
		return nil, nil
	}
	var defs []ParamDef
	if err := json.Unmarshal([]byte(paramsJson), &defs); err != nil {
		return nil, fmt.Errorf("params_json 不是合法的参数定义数组: %v", err)
	}
	seen := make(map[string]struct{}, len(defs))
	for i, d := range defs {
		if strings.TrimSpace(d.Name) == "" {
			return nil, fmt.Errorf("第 %d 个参数定义缺少 name", i+1)
		}
		switch d.Type {
		case TypeString, TypeNumber, TypeBool:
		default:
			return nil, fmt.Errorf("参数 %s 的类型 %q 非法，仅支持 string/number/bool", d.Name, d.Type)
		}
		if _, ok := seen[d.Name]; ok {
			return nil, fmt.Errorf("参数 %s 重复定义", d.Name)
		}
		seen[d.Name] = struct{}{}
	}
	return defs, nil
}

// coerceValue 按参数定义类型收敛 JSON 值（业务参数与 default 共用），类型不符即拒绝
func coerceValue(typ string, val interface{}) (interface{}, error) {
	switch typ {
	case TypeString:
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("期望 string 类型")
		}
		return s, nil
	case TypeNumber:
		n, ok := val.(float64) // encoding/json 的数字统一解析为 float64
		if !ok {
			return nil, fmt.Errorf("期望 number 类型")
		}
		return n, nil
	case TypeBool:
		b, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("期望 bool 类型")
		}
		return b, nil
	default:
		return nil, fmt.Errorf("不支持的参数类型 %q", typ)
	}
}

// ValidateParams 业务参数校验（纯函数）：
// 1) 业务参数必须是 JSON 对象；2) 缺 required 参数报错；
// 3) 白名单之外的“多余参数”拒绝；4) 非必填缺省时补 default。
// 返回 name -> 收敛后的参数值。
func ValidateParams(defs []ParamDef, paramsJson string) (map[string]interface{}, error) {
	provided := map[string]interface{}{}
	paramsJson = strings.TrimSpace(paramsJson)
	if paramsJson != "" {
		if err := json.Unmarshal([]byte(paramsJson), &provided); err != nil {
			return nil, fmt.Errorf("业务参数不是合法的 JSON 对象: %v", err)
		}
	}
	//白名单索引
	defIdx := make(map[string]ParamDef, len(defs))
	for _, d := range defs {
		defIdx[d.Name] = d
	}
	//多余参数拒绝（防注入、防误用）
	for name := range provided {
		if _, ok := defIdx[name]; !ok {
			return nil, fmt.Errorf("多余参数: %s", name)
		}
	}
	values := make(map[string]interface{}, len(defs))
	for _, d := range defs {
		if raw, ok := provided[d.Name]; ok {
			v, err := coerceValue(d.Type, raw)
			if err != nil {
				return nil, fmt.Errorf("参数 %s 类型错误: %v", d.Name, err)
			}
			values[d.Name] = v
			continue
		}
		//缺省：必填报错；非必填补 default（default 为 null/未配置则不传入 SQL）
		if d.Required {
			return nil, fmt.Errorf("缺少必填参数: %s", d.Name)
		}
		if d.Default != nil {
			v, err := coerceValue(d.Type, d.Default)
			if err != nil {
				return nil, fmt.Errorf("参数 %s 的 default 类型错误: %v", d.Name, err)
			}
			values[d.Name] = v
		}
	}
	return values, nil
}

// BuildSql 将 sql_text 中的 :name 占位符按出现顺序替换为 ? 并组装参数数组（纯函数）：
// 仅允许替换参数白名单内的占位符，白名单之外的占位符直接拒绝（防注入）。
// 注意：字符串字面量中的 :name 同样会被替换，属 :name 占位符方案的已知限制。
func BuildSql(sqlText string, defs []ParamDef, values map[string]interface{}) (string, []interface{}, error) {
	white := make(map[string]struct{}, len(defs))
	for _, d := range defs {
		white[d.Name] = struct{}{}
	}
	//先校验全部占位符都在白名单内
	for _, m := range paramNameRe.FindAllString(sqlText, -1) {
		if _, ok := white[m[1:]]; !ok {
			return "", nil, fmt.Errorf("SQL 中存在未声明的占位符 %s，已拒绝执行", m)
		}
	}
	//按出现顺序替换为 ? 并组装参数（未传值的可选参数以 NULL 传入）
	args := make([]interface{}, 0, len(defs))
	built := paramNameRe.ReplaceAllStringFunc(sqlText, func(m string) string {
		args = append(args, values[m[1:]])
		return "?"
	})
	return built, args, nil
}

// IsSelectStmt 判断语句是否以 SELECT 开头（剥掉前导括号后忽略大小写）；
// 写语句（INSERT/UPDATE/DELETE/DROP 等）返回 false，执行结果不缓存
func IsSelectStmt(sqlText string) bool {
	s := strings.TrimSpace(sqlText)
	s = strings.TrimSpace(strings.TrimPrefix(s, "(")) //支持 (select ...) 包装
	return len(s) >= 6 && strings.EqualFold(s[:6], "select")
}

// CacheKey 结果缓存 key：dsvc#api#{api_id}#{参数摘要}，
// 参数摘要对收敛后的参数做 sha256（map 序列化按键排序，结果稳定）
func CacheKey(apiId int64, values map[string]interface{}) string {
	raw, err := json.Marshal(values)
	if err != nil {
		raw = []byte(fmt.Sprintf("%v", values))
	}
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("dsvc#api#%d#%s", apiId, hex.EncodeToString(sum[:])[:16])
}

// MaskRows 按授权 field_mask 脱敏（纯函数）：结果集列名命中则值替换为 ****（列名忽略大小写）
func MaskRows(rows []map[string]interface{}, fieldMask string) {
	fieldMask = strings.TrimSpace(fieldMask)
	if fieldMask == "" || len(rows) == 0 {
		return
	}
	mask := make(map[string]struct{})
	for _, f := range strings.Split(fieldMask, ",") {
		f = strings.ToLower(strings.TrimSpace(f))
		if f != "" {
			mask[f] = struct{}{}
		}
	}
	if len(mask) == 0 {
		return
	}
	for _, row := range rows {
		for col := range row {
			if _, ok := mask[strings.ToLower(col)]; ok {
				row[col] = maskValue
			}
		}
	}
}

// normalizeWhitelist 规范化 IP 白名单：去首尾空白、丢弃空白项后逗号拼接
func normalizeWhitelist(whitelist string) string {
	items := strings.Split(whitelist, ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return strings.Join(out, ",")
}

// decodeRows 反序列化缓存中的结果集 JSON
func decodeRows(s string) ([]map[string]interface{}, error) {
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(s), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRowsAsMaps 在目标库执行 SELECT，把结果集转为通用行集；
// 驱动返回的 []byte（varchar/text/decimal）统一转 string，便于 JSON 输出
func QueryRowsAsMaps(ctx context.Context, db *sql.DB, query string, args []interface{}) ([]map[string]interface{}, error) {
	rs, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	cols, err := rs.Columns()
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]interface{}, 0, 16)
	for rs.Next() {
		container := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range container {
			ptrs[i] = &container[i]
		}
		if err := rs.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, c := range cols {
			row[c] = normalizeDbValue(container[i])
		}
		rows = append(rows, row)
	}
	return rows, rs.Err()
}

// normalizeDbValue 驱动返回值归一化（仅 []byte → string，其余保持原样）
func normalizeDbValue(v interface{}) interface{} {
	switch t := v.(type) {
	case []byte:
		return string(t)
	default:
		return v
	}
}

// ExecWrite 在目标库执行写语句，返回受影响行数
func ExecWrite(ctx context.Context, db *sql.DB, query string, args []interface{}) (int64, error) {
	ret, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}
