package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const domainTable = "`dc_theme_domain`"

// domainRows 查询列（显式 5 列，与 DomainRow 5 字段严格对齐；comment 为可空列，
// 用 ifnull 兜底避免 NULL 扫 string 报错。列数与扫描字段数不一致时 go-zero 报
// not matching destination to scan——少列扫宽 struct 同样报错，故专列窄结构体）
const domainRows = "`id`,`code`,`name`,IFNULL(`comment`,'') AS `comment`,`beid`"

// domainInsertCols 插入列清单（显式写出，不含自增 id 与 created_at，库默认值填充）
const domainInsertCols = "`code`,`name`,`comment`,`beid`"

// domainInsertFieldSize 插入字段数（与 domainInsertCols 严格对应）
const domainInsertFieldSize = 4

type (
	// DomainRow 主题域查询行：与 domainRows 的 5 列一一对应（不含时间列）
	DomainRow struct {
		Id      int64  `db:"id"`
		Code    string `db:"code"`
		Name    string `db:"name"`
		Comment string `db:"comment"`
		Beid    int64  `db:"beid"`
	}

	// DomainModel 数仓主题域（dc_theme_domain，uk_beid_code 唯一）
	DomainModel interface {
		// Insert 新增主题域，返回自增 id（beid+code 唯一，并发撞键归一为 ErrDupDomainCode）
		Insert(ctx context.Context, data *DomainRow) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*DomainRow, error)
		// FindByCode 按域编码查重（beid+code 唯一）
		FindByCode(ctx context.Context, beid int64, code string) (*DomainRow, error)
		// Update 全量更新（不含自增 id 与时间列，带多租户过滤）
		Update(ctx context.Context, data *DomainRow) error
		// Delete 按主键删除（是否可删由业务侧先查主题数决定）
		Delete(ctx context.Context, id, beid int64) error
		// FindList 全量列表（beid 必带，按 id 倒序；域数量少，不分页）
		FindList(ctx context.Context, beid int64) ([]*DomainRow, error)
	}

	defaultDomainModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewDomainModel 创建主题域 model
func NewDomainModel(conn sqlx.SqlConn) DomainModel {
	return &defaultDomainModel{
		conn:  conn,
		table: domainTable,
	}
}

// Insert 新增主题域，created_at 由库默认值填充
func (m *defaultDomainModel) Insert(ctx context.Context, data *DomainRow) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)", m.table, domainInsertCols, placeholders(domainInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query, data.Code, data.Name, data.Comment, data.Beid)
	if err != nil {
		//查重后并发窗口撞 uk_beid_code 时归一为业务错误
		if isDuplicateEntry(err) {
			return 0, ErrDupDomainCode
		}
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindOne 按主键查询主题域（beid>0 时追加多租户过滤）
func (m *defaultDomainModel) FindOne(ctx context.Context, id, beid int64) (*DomainRow, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", domainRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp DomainRow
	if err := queryOne(ctx, m.conn, &resp, query+" limit 1", args...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindByCode 按域编码查询（beid+code 唯一），供保存前查重
func (m *defaultDomainModel) FindByCode(ctx context.Context, beid int64, code string) (*DomainRow, error) {
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `code` = ? limit 1", domainRows, m.table)
	var resp DomainRow
	if err := queryOne(ctx, m.conn, &resp, query, beid, code); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update 全量更新主题域（编码唯一性由 uk_beid_code 兜底，冲突归一为业务错误）
func (m *defaultDomainModel) Update(ctx context.Context, data *DomainRow) error {
	query := fmt.Sprintf("update %s set `code` = ?, `name` = ?, `comment` = ? "+
		"where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.Code, data.Name, data.Comment, data.Id, data.Beid)
	if err != nil && isDuplicateEntry(err) {
		return ErrDupDomainCode
	}
	return err
}

// Delete 按主键删除主题域
func (m *defaultDomainModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 全量主题域列表（按 id 倒序）
func (m *defaultDomainModel) FindList(ctx context.Context, beid int64) ([]*DomainRow, error) {
	if err := ensureBeid(beid); err != nil {
		return nil, err
	}
	query := fmt.Sprintf("select %s from %s where `beid` = ? order by `id` desc", domainRows, m.table)
	var list []*DomainRow
	if err := m.conn.QueryRowsCtx(ctx, &list, query, beid); err != nil {
		return nil, err
	}
	return list, nil
}
