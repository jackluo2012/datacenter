package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	ruleTable           = "`dc_quality_rule`"
	ruleInsertFieldSize = 9 // 插入字段数（与 ruleInsertColumns 对应）
	// ruleInsertColumns 插入字段（不含自增 id 与时间列），与 Insert 传参顺序一一对应
	ruleInsertColumns = "name,rule_type,target_table,target_column,params_json,threshold_json,status,owner,beid"
	// ruleRows 查询列（params_json/threshold_json 允许 NULL，用 ifnull 兜底避免扫描报错；
	// 查询列数与 Rule 字段数一一对应，规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	ruleRows = "id,name,rule_type,target_table,target_column," +
		"ifnull(params_json,'') as params_json,ifnull(threshold_json,'') as threshold_json," +
		"status,owner,beid,created_at,updated_at"
)

// 质量规则（dc_quality_rule）
type (
	// Rule 质量规则行
	Rule struct {
		Id            int64     `db:"id"`
		Name          string    `db:"name"`
		RuleType      string    `db:"rule_type"` // row_count/not_null/unique/value_range/consistency/fluctuation
		TargetTable   string    `db:"target_table"`
		TargetColumn  string    `db:"target_column"`
		ParamsJson    string    `db:"params_json"`    // value_range={min,max}; consistency={ref_table,ref_column}; fluctuation={baseline,ratio_pct}
		ThresholdJson string    `db:"threshold_json"` // 期望阈值: {op:gt|gte|lt|lte|eq|between, value:..}（row_count 用）
		Status        int64     `db:"status"`         // 1.启用 0.停用
		Owner         string    `db:"owner"`
		Beid          int64     `db:"beid"`
		CreatedAt     time.Time `db:"created_at"`
		UpdatedAt     time.Time `db:"updated_at"`
	}

	// RuleModel 质量规则表操作接口
	RuleModel interface {
		// Insert 新增规则，返回自增 id
		Insert(ctx context.Context, data *Rule) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Rule, error)
		// Update 更新规则可编辑字段（按 id+beid 定位）
		Update(ctx context.Context, data *Rule) error
		// Delete 删除规则（按 id+beid 定位）
		Delete(ctx context.Context, id, beid int64) error
		// FindList 条件分页查询：beid 必带，ruleType 可选过滤；返回列表与总数
		FindList(ctx context.Context, beid int64, ruleType string, page, size int64) ([]*Rule, int64, error)
		// FindListByIds 按主键批量查询（RunJob 加载任务规则用，自动去重分片）
		FindListByIds(ctx context.Context, beid int64, ids []int64) ([]*Rule, error)
	}

	defaultRuleModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewRuleModel 创建质量规则 model
func NewRuleModel(conn sqlx.SqlConn) RuleModel {
	return &defaultRuleModel{
		conn:  conn,
		table: ruleTable,
	}
}

// Insert 新增质量规则
func (m *defaultRuleModel) Insert(ctx context.Context, data *Rule) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, ruleInsertColumns, placeholders(ruleInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Name, data.RuleType, data.TargetTable, data.TargetColumn,
		data.ParamsJson, data.ThresholdJson, data.Status, data.Owner, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询规则
func (m *defaultRuleModel) FindOne(ctx context.Context, id, beid int64) (*Rule, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", ruleRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Rule
	err := m.conn.QueryRowCtx(ctx, &resp, query+" limit 1", args...)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 更新规则（按 id+beid 定位，时间列由数据库自动维护）
func (m *defaultRuleModel) Update(ctx context.Context, data *Rule) error {
	query := fmt.Sprintf("update %s set `name` = ?, `rule_type` = ?, `target_table` = ?, `target_column` = ?, "+
		"`params_json` = ?, `threshold_json` = ?, `status` = ?, `owner` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Name, data.RuleType, data.TargetTable, data.TargetColumn,
		data.ParamsJson, data.ThresholdJson, data.Status, data.Owner,
		data.Id, data.Beid)
	return err
}

// Delete 删除规则
func (m *defaultRuleModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 规则列表（beid 多租户必带 + rule_type 可选过滤），返回列表与总数
func (m *defaultRuleModel) FindList(ctx context.Context, beid int64, ruleType string,
	page, size int64) ([]*Rule, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if ruleType != "" {
		where += " and `rule_type` = ?"
		args = append(args, ruleType)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		ruleRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Rule
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindListByIds 按主键批量查询规则（自动去重并分片，beid 多租户过滤）
func (m *defaultRuleModel) FindListByIds(ctx context.Context, beid int64, ids []int64) ([]*Rule, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	if len(ids) == 0 {
		return []*Rule{}, nil
	}
	uniq := dedupeIds(ids)
	list := make([]*Rule, 0, len(uniq))
	for start := 0; start < len(uniq); start += idQueryChunk {
		end := start + idQueryChunk
		if end > len(uniq) {
			end = len(uniq)
		}
		part := uniq[start:end]
		query := fmt.Sprintf("select %s from %s where `id` in (%s) and `beid` = ?",
			ruleRows, m.table, placeholders(len(part)))
		args := make([]any, 0, len(part)+1)
		for _, id := range part {
			args = append(args, id)
		}
		args = append(args, beid)
		var rows []*Rule
		if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
			return nil, err
		}
		list = append(list, rows...)
	}
	return list, nil
}
