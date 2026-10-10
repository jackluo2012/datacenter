package logic

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// 质量规则类型（dc_quality_rule.rule_type 白名单）
const (
	RuleTypeRowCount    = "row_count"   // 行数检查
	RuleTypeNotNull     = "not_null"    // 非空检查
	RuleTypeUnique      = "unique"      // 唯一性检查
	RuleTypeValueRange  = "value_range" // 值域检查
	RuleTypeConsistency = "consistency" // 一致性检查（外键引用完整性）
	RuleTypeFluctuation = "fluctuation" // 波动率检查（与基线行数比较）
)

// identifierRe 标识符白名单：表名/列名/库名仅允许字母、数字、下划线，
// 不通过者一律拒绝拼接（杜绝 `users; drop table x` 类注入）
var identifierRe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// ruleTypeSet 规则类型白名单集合
var ruleTypeSet = map[string]struct{}{
	RuleTypeRowCount:    {},
	RuleTypeNotNull:     {},
	RuleTypeUnique:      {},
	RuleTypeValueRange:  {},
	RuleTypeConsistency: {},
	RuleTypeFluctuation: {},
}

// ruleParams params_json 解析结果（按规则类型取用）
type ruleParams struct {
	// value_range
	Min, Max float64
	hasMin   bool
	hasMax   bool
	// consistency
	RefTable, RefColumn string
	// fluctuation
	Baseline, RatioPct float64
	hasBaseline        bool
	hasRatio           bool
}

// IsValidRuleType 规则类型是否在白名单内
func IsValidRuleType(ruleType string) bool {
	_, ok := ruleTypeSet[ruleType]
	return ok
}

// ValidIdentifier 标识符（库/表/列名）白名单校验，防 SQL 注入
func ValidIdentifier(name string) bool {
	return identifierRe.MatchString(name)
}

// QualifyTable 生成 `<db>.<table>` 全限定表名（db 为空时使用 DSN 默认库），
// 两侧标识符均先过白名单校验
func QualifyTable(db, table string) (string, error) {
	if !ValidIdentifier(table) {
		return "", fmt.Errorf("%w: 表名 %q", ErrInvalidIdentifier, table)
	}
	if db == "" {
		return table, nil
	}
	if !ValidIdentifier(db) {
		return "", fmt.Errorf("%w: 库名 %q", ErrInvalidIdentifier, db)
	}
	return db + "." + table, nil
}

// ValidateRule 创建规则前的静态校验：类型白名单 + 标识符白名单 +
// 按类型的 target_column/params_json 必填性（not_null/unique/value_range/consistency 需 target_column；
// value_range 需 {min,max}；consistency 需 {ref_table,ref_column}；fluctuation 需 {baseline,ratio_pct}）
func ValidateRule(ruleType, targetTable, targetColumn, paramsJson string) error {
	if !IsValidRuleType(ruleType) {
		return ErrInvalidRuleType
	}
	if strings.TrimSpace(targetTable) == "" {
		return ErrTargetTableRequired
	}
	if !ValidIdentifier(targetTable) {
		return fmt.Errorf("%w: target_table %q", ErrInvalidIdentifier, targetTable)
	}
	// not_null/unique/value_range/consistency 的检查 SQL 均需拼接目标列
	if ruleNeedsColumn(ruleType) && targetColumn == "" {
		return fmt.Errorf("%w: %s 需要 target_column", ErrInvalidRuleParams, ruleType)
	}
	if targetColumn != "" && !ValidIdentifier(targetColumn) {
		return fmt.Errorf("%w: target_column %q", ErrInvalidIdentifier, targetColumn)
	}

	switch ruleType {
	case RuleTypeValueRange:
		p, err := parseRuleParams(paramsJson)
		if err != nil {
			return err
		}
		if !p.hasMin || !p.hasMax {
			return fmt.Errorf("%w: value_range 需要 params_json 提供数字 min/max", ErrInvalidRuleParams)
		}
		if p.Min > p.Max {
			return fmt.Errorf("%w: value_range 的 min 不能大于 max", ErrInvalidRuleParams)
		}
	case RuleTypeConsistency:
		p, err := parseRuleParams(paramsJson)
		if err != nil {
			return err
		}
		if p.RefTable == "" || p.RefColumn == "" {
			return fmt.Errorf("%w: consistency 需要 params_json 提供 ref_table/ref_column", ErrInvalidRuleParams)
		}
		if !ValidIdentifier(p.RefTable) {
			return fmt.Errorf("%w: ref_table %q", ErrInvalidIdentifier, p.RefTable)
		}
		if !ValidIdentifier(p.RefColumn) {
			return fmt.Errorf("%w: ref_column %q", ErrInvalidIdentifier, p.RefColumn)
		}
	case RuleTypeFluctuation:
		if _, err := parseFluctuationParams(paramsJson); err != nil {
			return err
		}
	}
	return nil
}

// ruleNeedsColumn 该规则类型的检查 SQL 是否必须拼接目标列
func ruleNeedsColumn(ruleType string) bool {
	switch ruleType {
	case RuleTypeNotNull, RuleTypeUnique, RuleTypeValueRange, RuleTypeConsistency:
		return true
	default:
		return false
	}
}

// ParseRuleIds 解析任务配置的规则id串（逗号分隔）：去空白、去重、保持顺序，
// 空串或含非法片段报错
func ParseRuleIds(s string) ([]int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrEmptyRuleIds
	}
	seen := make(map[int64]struct{})
	ids := make([]int64, 0)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("%w: 片段 %q", ErrInvalidRuleIds, part)
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, ErrEmptyRuleIds
	}
	return ids, nil
}

// BuildCheckSQL 生成单条规则的检查 SQL（纯函数）。
// 表名/列名等标识符先过白名单校验才允许拼接；数值参数一律走占位符（返回 args）。
// db 为任务目标库（空则使用 DSN 默认库），非空时表名加 `<db>.` 前缀。
// 各类型 SQL：
//
//	row_count    → select count(*) from <table>
//	not_null     → select count(*) from <table> where <col> is null or <col>=''（实测 0 为通过）
//	unique       → select count(*) from (select <col> from <table> group by <col> having count(*)>1) t（实测 0 为通过）
//	value_range  → select count(*) from <table> where <col> not between min and max（实测 0 为通过）
//	consistency  → select count(*) from <table> t left join <ref_table> r on t.<col>=r.<ref_column>
//	               where r.<ref_column> is null（实测 0 为通过）
//	fluctuation  → select count(*) from <table>（与 baseline 比较，偏差 > ratio_pct% 不通过）
func BuildCheckSQL(ruleType, db, table, column, paramsJson string) (string, []any, error) {
	target, err := QualifyTable(db, table)
	if err != nil {
		return "", nil, err
	}

	// 需要拼列的类型先校验列名白名单
	if ruleNeedsColumn(ruleType) && !ValidIdentifier(column) {
		return "", nil, fmt.Errorf("%w: target_column %q", ErrInvalidIdentifier, column)
	}

	switch ruleType {
	case RuleTypeRowCount:
		return fmt.Sprintf("select count(*) from %s", target), nil, nil
	case RuleTypeNotNull:
		// null 或空串均计为脏数据
		return fmt.Sprintf("select count(*) from %s where `%s` is null or `%s` = ''", target, column, column), nil, nil
	case RuleTypeUnique:
		return fmt.Sprintf("select count(*) from (select `%s` from %s group by `%s` having count(*) > 1) t",
			column, target, column), nil, nil
	case RuleTypeValueRange:
		p, err := parseRuleParams(paramsJson)
		if err != nil {
			return "", nil, err
		}
		if !p.hasMin || !p.hasMax {
			return "", nil, fmt.Errorf("%w: value_range 需要 params_json 提供数字 min/max", ErrInvalidRuleParams)
		}
		// min/max 数值参数走占位符
		return fmt.Sprintf("select count(*) from %s where `%s` not between ? and ?", target, column),
			[]any{p.Min, p.Max}, nil
	case RuleTypeConsistency:
		p, err := parseRuleParams(paramsJson)
		if err != nil {
			return "", nil, err
		}
		refTable, err := QualifyTable(db, p.RefTable)
		if err != nil {
			return "", nil, fmt.Errorf("%w: ref_table %q", ErrInvalidIdentifier, p.RefTable)
		}
		if !ValidIdentifier(p.RefColumn) {
			return "", nil, fmt.Errorf("%w: ref_column %q", ErrInvalidIdentifier, p.RefColumn)
		}
		// 左连接无匹配（孤儿引用）计数，t/r 为固定别名字面量
		return fmt.Sprintf("select count(*) from %s t left join %s r on t.`%s` = r.`%s` where r.`%s` is null",
			target, refTable, column, p.RefColumn, p.RefColumn), nil, nil
	case RuleTypeFluctuation:
		// 当前实现取当前行数与基线比较，参数在判定阶段使用，这里提前校验缺参
		if _, err := parseFluctuationParams(paramsJson); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("select count(*) from %s", target), nil, nil
	default:
		return "", nil, ErrInvalidRuleType
	}
}

// JudgeCheck 检查结果判定（纯函数）：checkValue 为检查 SQL 实测值（命中数/行数），
// 返回是否通过与 detail_json（阈值/实测/差异对比详情）。
// 判定口径：
//
//	not_null/unique/value_range/consistency：命中脏数据 0 条即通过
//	fluctuation：|实测行数-基线|/基线 ≤ ratio_pct% 通过；基线为 0 时实测也为 0 才通过
//	row_count：threshold_json 提供阈值时按阈值比较，缺省要求表非空（>0）
func JudgeCheck(ruleType string, checkValue int64, paramsJson, thresholdJson string) (bool, string, error) {
	var (
		pass   bool
		detail map[string]any
		err    error
	)
	switch ruleType {
	case RuleTypeRowCount:
		pass, detail, err = judgeRowCount(checkValue, thresholdJson)
	case RuleTypeNotNull, RuleTypeUnique, RuleTypeConsistency:
		pass = checkValue == 0
		detail = map[string]any{"actual": checkValue, "expect": 0}
	case RuleTypeValueRange:
		pass = checkValue == 0
		detail = map[string]any{"actual": checkValue, "expect": 0}
		if p, perr := parseRuleParams(paramsJson); perr == nil && p.hasMin && p.hasMax {
			detail["min"] = p.Min
			detail["max"] = p.Max
		}
	case RuleTypeFluctuation:
		pass, detail, err = judgeFluctuation(checkValue, paramsJson)
	default:
		return false, "", ErrInvalidRuleType
	}
	if err != nil {
		return false, "", err
	}
	detail["pass"] = pass
	return pass, marshalDetail(detail), nil
}

// judgeRowCount 行数判定：threshold_json 为空时要求表非空；提供时按 {op,value} 阈值比较
func judgeRowCount(count int64, thresholdJson string) (bool, map[string]any, error) {
	detail := map[string]any{"actual": count}
	s := strings.TrimSpace(thresholdJson)
	if s == "" {
		// 缺省口径：表非空（>0）即通过
		return count > 0, detail, nil
	}
	var th struct {
		Op    string          `json:"op"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal([]byte(s), &th); err != nil {
		return false, nil, fmt.Errorf("%w: threshold_json 解析失败: %v", ErrInvalidRuleParams, err)
	}
	op := strings.ToLower(strings.TrimSpace(th.Op))
	values, err := parseThresholdValues(th.Value)
	if err != nil {
		return false, nil, err
	}
	detail["op"] = op
	detail["value"] = values
	return compareThreshold(count, op, values), detail, nil
}

// parseThresholdValues 解析阈值 value：单个数字或 between 的 [lo, hi] 数组
func parseThresholdValues(raw json.RawMessage) ([]float64, error) {
	var single float64
	if err := json.Unmarshal(raw, &single); err == nil {
		return []float64{single}, nil
	}
	var arr []float64
	if err := json.Unmarshal(raw, &arr); err != nil || len(arr) == 0 {
		return nil, fmt.Errorf("%w: threshold_json.value 需为数字或数字数组", ErrInvalidRuleParams)
	}
	return arr, nil
}

// compareThreshold 按阈值操作符比较实测行数（between 区间两端自动纠正为 lo<=hi）
func compareThreshold(count int64, op string, values []float64) bool {
	if len(values) == 0 {
		return false
	}
	v := float64(count)
	switch op {
	case "gt":
		return v > values[0]
	case "gte":
		return v >= values[0]
	case "lt":
		return v < values[0]
	case "lte":
		return v <= values[0]
	case "eq":
		return v == values[0]
	case "between":
		if len(values) < 2 {
			return false
		}
		lo, hi := values[0], values[1]
		if lo > hi {
			lo, hi = hi, lo
		}
		return v >= lo && v <= hi
	default:
		return false
	}
}

// judgeFluctuation 波动判定：|当前行数-基线|/基线*100 与 ratio_pct% 比较，超限不通过；
// 基线为 0 时数学上无法算偏差，约定当前行数也为 0 才通过（detail 里 deviation_pct 记 0）
func judgeFluctuation(current int64, paramsJson string) (bool, map[string]any, error) {
	p, err := parseFluctuationParams(paramsJson)
	if err != nil {
		return false, nil, err
	}
	detail := map[string]any{
		"actual":    current,
		"baseline":  p.Baseline,
		"ratio_pct": p.RatioPct,
	}
	var deviation float64
	pass := false
	switch {
	case p.Baseline > 0:
		deviation = math.Abs(float64(current)-p.Baseline) / p.Baseline * 100
		pass = deviation <= p.RatioPct
	default:
		// 基线 <= 0：无偏差可算，仅当前也为 0 时视为无波动
		pass = current == 0
	}
	detail["deviation_pct"] = deviation
	return pass, detail, nil
}

// parseFluctuationParams 解析并校验 fluctuation 参数（baseline/ratio_pct 必须为数字）
func parseFluctuationParams(paramsJson string) (*ruleParams, error) {
	p, err := parseRuleParams(paramsJson)
	if err != nil {
		return nil, err
	}
	if !p.hasBaseline || !p.hasRatio {
		return nil, fmt.Errorf("%w: fluctuation 需要 params_json 提供数字 baseline/ratio_pct", ErrInvalidRuleParams)
	}
	if p.RatioPct < 0 {
		return nil, fmt.Errorf("%w: fluctuation 的 ratio_pct 不能为负数", ErrInvalidRuleParams)
	}
	return p, nil
}

// parseRuleParams 解析 params_json（空串视为空对象），宽容兼容 json 数值与字符串数字
func parseRuleParams(paramsJson string) (*ruleParams, error) {
	m, err := parseParamsJson(paramsJson)
	if err != nil {
		return nil, err
	}
	p := &ruleParams{}
	p.Min, p.hasMin = paramNumber(m, "min")
	p.Max, p.hasMax = paramNumber(m, "max")
	p.RefTable, _ = paramString(m, "ref_table")
	p.RefColumn, _ = paramString(m, "ref_column")
	p.Baseline, p.hasBaseline = paramNumber(m, "baseline")
	p.RatioPct, p.hasRatio = paramNumber(m, "ratio_pct")
	return p, nil
}

// parseParamsJson 解析 params_json 为 map（空串/空白视为空对象）
func parseParamsJson(paramsJson string) (map[string]any, error) {
	s := strings.TrimSpace(paramsJson)
	if s == "" {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, fmt.Errorf("%w: params_json 解析失败: %v", ErrInvalidRuleParams, err)
	}
	return m, nil
}

// paramNumber 取数字参数：兼容 json 数值 / 字符串数字 / json.Number
func paramNumber(m map[string]any, key string) (float64, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// paramString 取字符串参数
func paramString(m map[string]any, key string) (string, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// marshalDetail 详情 map 序列化为 json 串（序列化失败降级为空串，不阻断主流程）
func marshalDetail(m map[string]any) string {
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}
