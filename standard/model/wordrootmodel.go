package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	wordRootTable           = "`dc_std_word_root`"
	wordRootInsertFieldSize = 5 // 插入字段数（与 wordRootInsertColumns 对应）
	// wordRootInsertColumns 插入字段（不含自增 id 与时间列），与 Insert 传参顺序一一对应
	wordRootInsertColumns = "`name`,`cn_name`,`type`,`comment`,`beid`"
	// wordRootRows 查询列（8 列与 WordRoot 8 字段一一对应，
	// 规避 go-zero 少列查询扫宽 struct 报 not matching destination to scan）
	wordRootRows = "`id`,`name`,`cn_name`,`type`,`comment`,`beid`,`created_at`,`updated_at`"
)

// 标准词根（dc_std_word_root）
type (
	// WordRoot 词根行
	WordRoot struct {
		Id        int64     `db:"id"`
		Name      string    `db:"name"`
		CnName    string    `db:"cn_name"`
		Type      string    `db:"type"` // common/attribute/dimension/indicator
		Comment   string    `db:"comment"`
		Beid      int64     `db:"beid"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	// WordRootModel 词根表操作接口
	WordRootModel interface {
		// Insert 新增词根，返回自增 id
		Insert(ctx context.Context, data *WordRoot) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*WordRoot, error)
		// FindOneByName 按 beid + 词根名查询（保存前查重用）
		FindOneByName(ctx context.Context, beid int64, name string) (*WordRoot, error)
		// Update 全量更新词根可编辑字段（按 id+beid 定位）
		Update(ctx context.Context, data *WordRoot) error
		// Delete 删除词根（按 id+beid 定位）
		Delete(ctx context.Context, id, beid int64) error
		// FindList 条件分页查询：beid 必带，keyword 对 name/cn_name 模糊匹配；返回列表与总数
		FindList(ctx context.Context, beid int64, keyword string, page, size int64) ([]*WordRoot, int64, error)
	}

	defaultWordRootModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewWordRootModel 创建词根 model
func NewWordRootModel(conn sqlx.SqlConn) WordRootModel {
	return &defaultWordRootModel{
		conn:  conn,
		table: wordRootTable,
	}
}

// Insert 新增词根
func (m *defaultWordRootModel) Insert(ctx context.Context, data *WordRoot) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, wordRootInsertColumns, placeholders(wordRootInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Name, data.CnName, data.Type, data.Comment, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询词根
func (m *defaultWordRootModel) FindOne(ctx context.Context, id, beid int64) (*WordRoot, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", wordRootRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp WordRoot
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

// FindOneByName 按 beid + 词根名查询（uk_beid_name 查重用）
func (m *defaultWordRootModel) FindOneByName(ctx context.Context, beid int64, name string) (*WordRoot, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `name` = ? limit 1", wordRootRows, m.table)
	var resp WordRoot
	err := m.conn.QueryRowCtx(ctx, &resp, query, beid, name)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 更新词根（按 id+beid 定位，时间列由数据库自动维护）
func (m *defaultWordRootModel) Update(ctx context.Context, data *WordRoot) error {
	query := fmt.Sprintf("update %s set `name` = ?, `cn_name` = ?, `type` = ?, `comment` = ? "+
		"where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Name, data.CnName, data.Type, data.Comment, data.Id, data.Beid)
	return err
}

// Delete 删除词根
func (m *defaultWordRootModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 词根列表（beid 多租户必带 + keyword 对 name/cn_name 模糊匹配），返回列表与总数
func (m *defaultWordRootModel) FindList(ctx context.Context, beid int64, keyword string,
	page, size int64) ([]*WordRoot, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and (`name` like ? or `cn_name` like ?)"
		like := "%" + escapeLike(kw) + "%"
		args = append(args, like, like)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		wordRootRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*WordRoot
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
