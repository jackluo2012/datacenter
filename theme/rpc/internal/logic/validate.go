package logic

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// validate.go 纯函数校验：layer 白名单、fields_json 字段定义、dim_json 维度声明、
// select_sql 视图查询语句、refresh_mode 刷新模式，以及各类标识符白名单

var (
	// identifierRe 物理表名/视图名/域编码等标识符：小写字母开头，仅小写字母/数字/下划线
	// （表名/视图名会真实建表建视图，严格防注入）
	identifierRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	// fieldNameRe 字段名：字母开头，仅字母/数字/下划线（进 DDL 前的列名白名单）
	fieldNameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
)

// 数仓分层白名单（dc_theme_model.layer，物理表名以 <layer>_<table> 落地）
const (
	LayerOds = "ods"
	LayerDwd = "dwd"
	LayerDws = "dws"
	LayerAds = "ads"
	LayerDim = "dim"
)

// layers 分层白名单集合
var layers = map[string]struct{}{
	LayerOds: {},
	LayerDwd: {},
	LayerDws: {},
	LayerAds: {},
	LayerDim: {},
}

// 刷新模式白名单（dc_theme_mv.refresh_mode，schedule 为定时刷新存档标记）
const (
	RefreshModeManual   = "manual"
	RefreshModeSchedule = "schedule"
)

// 物化视图刷新结果（dc_theme_mv.last_status）
const (
	MvStatusSuccess = "success"
	MvStatusFailed  = "failed"
)

// 标识符长度上限（MySQL 标识符 64 字符，表/视图/库名共用）
const identifierMaxLen = 64

// FieldDef 数仓模型字段定义（fields_json 数组元素，对齐 proto 注释
// [{name,type,comment,pk,dim}]）：name/type 必填，pk 标注主键（整组必须恰有一个），
// dim 标注是否维度退化字段
type FieldDef struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Comment string `json:"comment,omitempty"`
	Pk      bool   `json:"pk,omitempty"`
	Dim     bool   `json:"dim,omitempty"`
}

// DimDef 维度声明（dim_json 数组元素 [{dim_table,join_key}]）
type DimDef struct {
	DimTable string `json:"dim_table"`
	JoinKey  string `json:"join_key"`
}

// ValidLayer 数仓分层是否在白名单
func ValidLayer(layer string) bool {
	_, ok := layers[layer]
	return ok
}

// ValidIdentifier 物理表名/视图名/编码是否合法（小写字母开头，仅小写字母/数字/下划线）
func ValidIdentifier(s string) bool {
	return identifierRe.MatchString(s)
}

// ValidFieldName 字段名是否合法（字母开头，仅字母/数字/下划线）
func ValidFieldName(s string) bool {
	return fieldNameRe.MatchString(s)
}

// ValidRefreshMode 刷新模式是否在白名单
func ValidRefreshMode(s string) bool {
	return s == RefreshModeManual || s == RefreshModeSchedule
}

// trimSpace 去首尾空白
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// ParseFieldsJson 解析已存储的模型字段定义（保存时已校验，此处防御性解析）
func ParseFieldsJson(fieldsJson string) ([]FieldDef, error) {
	var fields []FieldDef
	if err := json.Unmarshal([]byte(fieldsJson), &fields); err != nil {
		return nil, fmt.Errorf("%w: 须为 JSON 数组: %v", ErrInvalidFieldsJson, err)
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

// ValidateFieldsJson 校验模型字段定义 JSON：
//   - 须为 JSON 数组且非空；
//   - 每项含非空合法的 name（字母开头，仅字母/数字/下划线，防注入）与白名单内的 type
//     （string/int/bigint/decimal/bool/date/datetime）；
//   - name 不得重复；
//   - 主键（pk=true）必须恰有一个。
//
// 返回解析后的字段列表（comment/dim 附加项合法保留）
func ValidateFieldsJson(fieldsJson string) ([]FieldDef, error) {
	var fields []FieldDef
	if err := json.Unmarshal([]byte(fieldsJson), &fields); err != nil {
		return nil, fmt.Errorf("%w: 须为 JSON 数组: %v", ErrInvalidFieldsJson, err)
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("%w: 字段定义不能为空", ErrInvalidFieldsJson)
	}
	seen := make(map[string]struct{}, len(fields))
	pkCount := 0
	for i, f := range fields {
		if strings.TrimSpace(f.Name) == "" {
			return nil, fmt.Errorf("%w: 第 %d 个字段缺少 name", ErrInvalidFieldsJson, i+1)
		}
		if !ValidFieldName(f.Name) {
			return nil, fmt.Errorf("%w: 字段名 %s 不合法（须以字母开头，仅含字母/数字/下划线）",
				ErrInvalidFieldsJson, f.Name)
		}
		if _, dup := seen[f.Name]; dup {
			return nil, fmt.Errorf("%w: 字段 %s 重复定义", ErrInvalidFieldsJson, f.Name)
		}
		seen[f.Name] = struct{}{}
		if !ValidFieldType(f.Type) {
			return nil, fmt.Errorf("%w: 字段 %s 类型 %s 不在白名单（string/int/bigint/decimal/bool/date/datetime）",
				ErrInvalidFieldsJson, f.Name, f.Type)
		}
		if f.Pk {
			pkCount++
		}
	}
	//pk 断言：主键必须恰有一个（建表 primary key 子句依赖）
	if pkCount != 1 {
		return nil, fmt.Errorf("%w: 主键字段（pk）必须恰有一个，当前 %d 个", ErrInvalidFieldsJson, pkCount)
	}
	return fields, nil
}

// ValidateDimJson 校验维度声明（可选）：空串直接跳过；非空须为 [{dim_table,join_key}]
// JSON 数组且两项均非空。返回规范化后的 dim_json（空串原样返回）
func ValidateDimJson(dimJson string) (string, error) {
	s := trimSpace(dimJson)
	if s == "" {
		return "", nil
	}
	var dims []DimDef
	if err := json.Unmarshal([]byte(s), &dims); err != nil {
		return "", fmt.Errorf("%w: 须为 [{dim_table,join_key}] JSON 数组: %v", ErrInvalidDimJson, err)
	}
	for i, d := range dims {
		if strings.TrimSpace(d.DimTable) == "" || strings.TrimSpace(d.JoinKey) == "" {
			return "", fmt.Errorf("%w: 第 %d 项 dim_table/join_key 不能为空", ErrInvalidDimJson, i+1)
		}
	}
	b, err := json.Marshal(dims)
	if err != nil {
		return "", fmt.Errorf("%w: 序列化失败: %v", ErrInvalidDimJson, err)
	}
	return string(b), nil
}

// ValidateSelectSql 物化视图 select_sql 校验（纯函数）：去空白后必须以 SELECT 开头
// （大小写不敏感），且不含分号（防多语句注入——执行时与 create view 拼接）
func ValidateSelectSql(selectSql string) error {
	s := trimSpace(selectSql)
	if s == "" || !strings.HasPrefix(strings.ToLower(s), "select") {
		return ErrInvalidSelectSql
	}
	if strings.Contains(s, ";") {
		return ErrInvalidSelectSql
	}
	return nil
}
