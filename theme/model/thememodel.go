package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const themeTable = "`dc_theme`"

// themeRows 查询列（显式 6 列，与 ThemeRow 6 字段严格对齐；comment 为可空列，
// 用 ifnull 兜底避免 NULL 扫 string 报错）
const themeRows = "`id`,`domain_id`,`code`,`name`,IFNULL(`comment`,'') AS `comment`,`beid`"

// themeInsertCols 插入列清单（不含自增 id 与 created_at）
const themeInsertCols = "`domain_id`,`code`,`name`,`comment`,`beid`"

// themeInsertFieldSize 插入字段数（与 themeInsertCols 严格对应）
const themeInsertFieldSize = 5

type (
	// ThemeRow 主题查询行：与 themeRows 的 6 列一一对应（不含时间列）
	ThemeRow struct {
		Id       int64  `db:"id"`
		DomainId int64  `db:"domain_id"`
		Code     string `db:"code"`
		Name     string `db:"name"`
		Comment  string `db:"comment"`
		Beid     int64  `db:"beid"`
	}

	// ThemeModel 数仓主题（dc_theme，uk_beid_code 唯一）
	ThemeModel interface {
		// Insert 新增主题，返回自增 id（beid+code 唯一，并发撞键归一为 ErrDupThemeCode）
		Insert(ctx context.Context, data *ThemeRow) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*ThemeRow, error)
		// FindByCode 按主题编码查重（beid+code 唯一）
		FindByCode(ctx context.Context, beid int64, code string) (*ThemeRow, error)
		// Update 全量更新（不含自增 id 与时间列，带多租户过滤）
		Update(ctx context.Context, data *ThemeRow) error
		// Delete 按主键删除（是否可删由业务侧先查模型数决定）
		Delete(ctx context.Context, id, beid int64) error
		// FindList 条件分页（beid 必带，domainId>0 时按域过滤）
		FindList(ctx context.Context, beid, domainId, page, size int64) ([]*ThemeRow, int64, error)
		// CountByDomain 统计域下主题数（删除域前校验用）
		CountByDomain(ctx context.Context, domainId int64) (int64, error)
	}

	defaultThemeModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewThemeModel 创建主题 model
func NewThemeModel(conn sqlx.SqlConn) ThemeModel {
	return &defaultThemeModel{
		conn:  conn,
		table: themeTable,
	}
}

// Insert 新增主题，created_at 由库默认值填充
func (m *defaultThemeModel) Insert(ctx context.Context, data *ThemeRow) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)", m.table, themeInsertCols, placeholders(themeInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query, data.DomainId, data.Code, data.Name, data.Comment, data.Beid)
	if err != nil {
		if isDuplicateEntry(err) {
			return 0, ErrDupThemeCode
		}
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindOne 按主键查询主题（beid>0 时追加多租户过滤）
func (m *defaultThemeModel) FindOne(ctx context.Context, id, beid int64) (*ThemeRow, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", themeRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp ThemeRow
	if err := queryOne(ctx, m.conn, &resp, query+" limit 1", args...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindByCode 按主题编码查询（beid+code 唯一），供保存前查重
func (m *defaultThemeModel) FindByCode(ctx context.Context, beid int64, code string) (*ThemeRow, error) {
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `code` = ? limit 1", themeRows, m.table)
	var resp ThemeRow
	if err := queryOne(ctx, m.conn, &resp, query, beid, code); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update 全量更新主题（编码唯一性由 uk_beid_code 兜底，冲突归一为业务错误）
func (m *defaultThemeModel) Update(ctx context.Context, data *ThemeRow) error {
	query := fmt.Sprintf("update %s set `domain_id` = ?, `code` = ?, `name` = ?, `comment` = ? "+
		"where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.DomainId, data.Code, data.Name, data.Comment, data.Id, data.Beid)
	if err != nil && isDuplicateEntry(err) {
		return ErrDupThemeCode
	}
	return err
}

// Delete 按主键删除主题
func (m *defaultThemeModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 条件分页查询主题，domainId>0 时按域过滤，返回列表与总数
func (m *defaultThemeModel) FindList(ctx context.Context, beid, domainId, page, size int64) ([]*ThemeRow, int64, error) {
	if err := ensureBeid(beid); err != nil {
		return nil, 0, err
	}
	where := "`beid` = ?"
	args := []any{beid}
	if domainId > 0 {
		where += " and `domain_id` = ?"
		args = append(args, domainId)
	}

	//总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	//分页
	page, size = normalizePage(page, size)
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?", themeRows, m.table, where)
	args = append(args, size, (page-1)*size)
	var list []*ThemeRow
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// CountByDomain 统计指定域下的主题数
func (m *defaultThemeModel) CountByDomain(ctx context.Context, domainId int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `domain_id` = ?", m.table)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, domainId); err != nil {
		return 0, err
	}
	return total, nil
}
