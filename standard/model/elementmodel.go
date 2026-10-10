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
	elementTable           = "`dc_std_element`"
	elementInsertFieldSize = 8 // 插入字段数（与 elementInsertColumns 对应）
	// elementInsertColumns 插入字段（不含自增 id 与时间列），与 Insert 传参顺序一一对应
	elementInsertColumns = "`element_code`,`element_name`,`data_type`,`data_length`,`data_precision`,`dict_code`,`comment`,`beid`"
	// elementRows 查询列（11 列与 Element 11 字段一一对应，
	// 规避 go-zero 少列查询扫宽 struct 报 not matching destination to scan）
	elementRows = "`id`,`element_code`,`element_name`,`data_type`,`data_length`,`data_precision`," +
		"ifnull(`dict_code`,'') as `dict_code`,`comment`,`beid`,`created_at`,`updated_at`"
)

// 数据元（dc_std_element）
type (
	// Element 数据元行
	Element struct {
		Id            int64     `db:"id"`
		ElementCode   string    `db:"element_code"`
		ElementName   string    `db:"element_name"`
		DataType      string    `db:"data_type"` // varchar/int/bigint/decimal/date/datetime/text
		DataLength    int32     `db:"data_length"`
		DataPrecision int32     `db:"data_precision"` // 小数位（decimal 用）
		DictCode      string    `db:"dict_code"`      // 值域引用字典编码，空为无
		Comment       string    `db:"comment"`
		Beid          int64     `db:"beid"`
		CreatedAt     time.Time `db:"created_at"`
		UpdatedAt     time.Time `db:"updated_at"`
	}

	// ElementModel 数据元表操作接口
	ElementModel interface {
		// Insert 新增数据元，返回自增 id
		Insert(ctx context.Context, data *Element) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Element, error)
		// FindOneByCode 按 beid + 数据元编码查询（保存前查重用）
		FindOneByCode(ctx context.Context, beid int64, elementCode string) (*Element, error)
		// Update 更新数据元可编辑字段（按 id+beid 定位）
		Update(ctx context.Context, data *Element) error
		// Delete 删除数据元（按 id+beid 定位）
		Delete(ctx context.Context, id, beid int64) error
		// FindList 条件分页查询：beid 必带，keyword 对 element_code/element_name 模糊匹配；返回列表与总数
		FindList(ctx context.Context, beid int64, keyword string, page, size int64) ([]*Element, int64, error)
		// FindListByCodes 按 beid + 编码集合批量查询（标准模型 elements_json 校验/快照组装用），自动去重分片
		FindListByCodes(ctx context.Context, beid int64, codes []string) ([]*Element, error)
	}

	defaultElementModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewElementModel 创建数据元 model
func NewElementModel(conn sqlx.SqlConn) ElementModel {
	return &defaultElementModel{
		conn:  conn,
		table: elementTable,
	}
}

// Insert 新增数据元
func (m *defaultElementModel) Insert(ctx context.Context, data *Element) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, elementInsertColumns, placeholders(elementInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query,
		data.ElementCode, data.ElementName, data.DataType, data.DataLength,
		data.DataPrecision, data.DictCode, data.Comment, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询数据元
func (m *defaultElementModel) FindOne(ctx context.Context, id, beid int64) (*Element, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", elementRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Element
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

// FindOneByCode 按 beid + 数据元编码查询
func (m *defaultElementModel) FindOneByCode(ctx context.Context, beid int64, elementCode string) (*Element, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `element_code` = ? limit 1",
		elementRows, m.table)
	var resp Element
	err := m.conn.QueryRowCtx(ctx, &resp, query, beid, elementCode)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 更新数据元（按 id+beid 定位，时间列由数据库自动维护）
func (m *defaultElementModel) Update(ctx context.Context, data *Element) error {
	query := fmt.Sprintf("update %s set `element_code` = ?, `element_name` = ?, `data_type` = ?, "+
		"`data_length` = ?, `data_precision` = ?, `dict_code` = ?, `comment` = ? "+
		"where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		data.ElementCode, data.ElementName, data.DataType, data.DataLength,
		data.DataPrecision, data.DictCode, data.Comment, data.Id, data.Beid)
	return err
}

// Delete 删除数据元
func (m *defaultElementModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 数据元列表（beid 多租户必带 + keyword 对 element_code/element_name 模糊匹配），返回列表与总数
func (m *defaultElementModel) FindList(ctx context.Context, beid int64, keyword string,
	page, size int64) ([]*Element, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and (`element_code` like ? or `element_name` like ?)"
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
		elementRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Element
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindListByCodes 按编码集合批量查询数据元（自动去重并分片，beid 多租户过滤）
func (m *defaultElementModel) FindListByCodes(ctx context.Context, beid int64, codes []string) ([]*Element, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	uniq := dedupeStrings(codes)
	if len(uniq) == 0 {
		return []*Element{}, nil
	}
	list := make([]*Element, 0, len(uniq))
	for start := 0; start < len(uniq); start += codeQueryChunk {
		end := start + codeQueryChunk
		if end > len(uniq) {
			end = len(uniq)
		}
		part := uniq[start:end]
		query := fmt.Sprintf("select %s from %s where `beid` = ? and `element_code` in (%s)",
			elementRows, m.table, placeholders(len(part)))
		args := make([]any, 0, len(part)+1)
		args = append(args, beid)
		for _, code := range part {
			args = append(args, code)
		}
		var rows []*Element
		if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
			return nil, err
		}
		list = append(list, rows...)
	}
	return list, nil
}
