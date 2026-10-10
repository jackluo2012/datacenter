package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	entityTagTable = "`dc_sec_entity_tag`"
	// entityTagInsertColumns 插入字段（不含自增 id 与时间列，updated_at 由数据库自动维护）
	entityTagInsertColumns = "`entity_id`,`entity_name`,`level_code`,`mask_rule_code`,`beid`"
	// entityTagRows 查询列（6 列与 EntityTag 6 字段一一对应，规避扫描错位）
	entityTagRows = "`id`,`entity_id`,`entity_name`,`level_code`,`mask_rule_code`,`beid`"
)

// 实体分级打标（dc_sec_entity_tag）：元数据实体 ← 分级，可绑定脱敏规则
type (
	// EntityTag 打标行
	EntityTag struct {
		Id           int64  `db:"id"`
		EntityId     int64  `db:"entity_id"`
		EntityName   string `db:"entity_name"`
		LevelCode    string `db:"level_code"`     //分级编码
		MaskRuleCode string `db:"mask_rule_code"` //绑定脱敏规则，空为未绑定
		Beid         int64  `db:"beid"`
	}

	// EntityTagModel 实体打标表操作接口
	EntityTagModel interface {
		// TagEntity 按 uk_beid_entity upsert（同实体重复打标等价覆盖），返回记录 id
		TagEntity(ctx context.Context, data *EntityTag) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*EntityTag, error)
		// List 条件分页：beid 必带，levelCode 非空时按密级过滤；返回列表与总数
		List(ctx context.Context, beid int64, levelCode string, page, size int64) ([]*EntityTag, int64, error)
		// CountByRuleCode 统计引用某脱敏规则的打标数（规则删除前的引用检查）
		CountByRuleCode(ctx context.Context, beid int64, ruleCode string) (int64, error)
	}

	defaultEntityTagModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewEntityTagModel 创建实体打标 model
func NewEntityTagModel(conn sqlx.SqlConn) EntityTagModel {
	return &defaultEntityTagModel{
		conn:  conn,
		table: entityTagTable,
	}
}

// TagEntity 按 uk_beid_entity upsert 打标记录。
// 更新分支借 `id = last_insert_id(id)` 让 LastInsertId 返回既有主键（insert/update 两分支都可靠）
func (m *defaultEntityTagModel) TagEntity(ctx context.Context, data *EntityTag) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?) "+
		"on duplicate key update `id` = last_insert_id(`id`), `entity_name` = values(`entity_name`), "+
		"`level_code` = values(`level_code`), `mask_rule_code` = values(`mask_rule_code`)",
		m.table, entityTagInsertColumns)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.EntityId, data.EntityName, data.LevelCode, data.MaskRuleCode, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询打标记录
func (m *defaultEntityTagModel) FindOne(ctx context.Context, id, beid int64) (*EntityTag, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", entityTagRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp EntityTag
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

// List 打标列表（beid 多租户必带，levelCode 非空时按密级过滤），返回列表与总数
func (m *defaultEntityTagModel) List(ctx context.Context, beid int64, levelCode string,
	page, size int64) ([]*EntityTag, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if levelCode != "" {
		where += " and `level_code` = ?"
		args = append(args, levelCode)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		entityTagRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*EntityTag
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// CountByRuleCode 统计绑定某脱敏规则的打标数（>0 时规则禁止删除）
func (m *defaultEntityTagModel) CountByRuleCode(ctx context.Context, beid int64, ruleCode string) (int64, error) {
	if beid <= 0 {
		return 0, ErrBeidRequired
	}
	query := fmt.Sprintf("select count(*) from %s where `beid` = ? and `mask_rule_code` = ?", m.table)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, beid, ruleCode); err != nil {
		return 0, err
	}
	return total, nil
}
