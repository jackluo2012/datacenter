package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	maskRuleTable = "`dc_sec_mask_rule`"
	// maskRuleInsertColumns 插入字段（不含自增 id，created_at 由数据库默认当前时间）
	maskRuleInsertColumns = "`rule_code`,`rule_name`,`algorithm`,`params_json`,`beid`"
	// maskRuleRows 查询列（params_json 为可空 text 列，用 ifnull 兜底为空串，
	// 6 列与 MaskRule 6 字段一一对应，规避扫描错位）
	maskRuleRows = "`id`,`rule_code`,`rule_name`,`algorithm`,ifnull(`params_json`,'') as `params_json`,`beid`"
)

// 脱敏规则库（dc_sec_mask_rule）
type (
	// MaskRule 脱敏规则行
	MaskRule struct {
		Id         int64  `db:"id"`
		RuleCode   string `db:"rule_code"`   //如 mask_phone
		RuleName   string `db:"rule_name"`   //
		Algorithm  string `db:"algorithm"`   //mask/hash/replace/null
		ParamsJson string `db:"params_json"` //算法参数：{"prefix":3,"suffix":2,"mask_char":"*"} / {"replace":""}
		Beid       int64  `db:"beid"`
	}

	// MaskRuleModel 脱敏规则表操作接口
	MaskRuleModel interface {
		// Insert 新增规则，返回自增 id（编码查重由 logic 层完成）
		Insert(ctx context.Context, data *MaskRule) (int64, error)
		// Update 更新规则可编辑字段（按 id+beid 定位）
		Update(ctx context.Context, data *MaskRule) error
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*MaskRule, error)
		// FindByCode 按 beid + rule_code 查询（保存查重/打标引用校验用）
		FindByCode(ctx context.Context, beid int64, ruleCode string) (*MaskRule, error)
		// Delete 删除规则（按 id+beid 定位）
		Delete(ctx context.Context, id, beid int64) error
		// List 条件分页查询：beid 必带；返回列表与总数
		List(ctx context.Context, beid int64, page, size int64) ([]*MaskRule, int64, error)
	}

	defaultMaskRuleModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewMaskRuleModel 创建脱敏规则 model
func NewMaskRuleModel(conn sqlx.SqlConn) MaskRuleModel {
	return &defaultMaskRuleModel{
		conn:  conn,
		table: maskRuleTable,
	}
}

// Insert 新增脱敏规则
func (m *defaultMaskRuleModel) Insert(ctx context.Context, data *MaskRule) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?)",
		m.table, maskRuleInsertColumns)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.RuleCode, data.RuleName, data.Algorithm, data.ParamsJson, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// Update 更新脱敏规则（按 id+beid 定位）
func (m *defaultMaskRuleModel) Update(ctx context.Context, data *MaskRule) error {
	query := fmt.Sprintf("update %s set `rule_code` = ?, `rule_name` = ?, `algorithm` = ?, "+
		"`params_json` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		data.RuleCode, data.RuleName, data.Algorithm, data.ParamsJson, data.Id, data.Beid)
	return err
}

// FindOne 按主键查询脱敏规则
func (m *defaultMaskRuleModel) FindOne(ctx context.Context, id, beid int64) (*MaskRule, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", maskRuleRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp MaskRule
	err := m.conn.QueryRowCtx(ctx, &resp, query+" limit 1", args...)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByCode 按 beid + 规则编码查询
func (m *defaultMaskRuleModel) FindByCode(ctx context.Context, beid int64, ruleCode string) (*MaskRule, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `rule_code` = ? limit 1",
		maskRuleRows, m.table)
	var resp MaskRule
	err := m.conn.QueryRowCtx(ctx, &resp, query, beid, ruleCode)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Delete 删除脱敏规则
func (m *defaultMaskRuleModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// List 脱敏规则列表（beid 多租户必带），返回列表与总数
func (m *defaultMaskRuleModel) List(ctx context.Context, beid int64,
	page, size int64) ([]*MaskRule, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where `beid` = ?", m.table)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, beid); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where `beid` = ? order by `id` desc limit ?, ?",
		maskRuleRows, m.table)
	var list []*MaskRule
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, beid, (page-1)*size, size); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
