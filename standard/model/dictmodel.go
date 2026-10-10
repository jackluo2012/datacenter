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
	dictTable           = "`dc_std_dict`"
	dictInsertFieldSize = 5 // 插入字段数（与 dictInsertColumns 对应）
	// dictInsertColumns 插入字段（不含自增 id 与时间列），与 Insert 传参顺序一一对应
	dictInsertColumns = "`dict_code`,`dict_name`,`comment`,`status`,`beid`"
	// dictRows 查询列（8 列与 Dict 8 字段一一对应，
	// 规避 go-zero 少列查询扫宽 struct 报 not matching destination to scan）
	dictRows = "`id`,`dict_code`,`dict_name`,`comment`,`status`,`beid`,`created_at`,`updated_at`"
)

// 标准字典（dc_std_dict）
type (
	// Dict 字典行
	Dict struct {
		Id        int64     `db:"id"`
		DictCode  string    `db:"dict_code"`
		DictName  string    `db:"dict_name"`
		Comment   string    `db:"comment"`
		Status    int64     `db:"status"` // 1.启用 0.停用
		Beid      int64     `db:"beid"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	// DictModel 字典表操作接口
	// 写方法带 session 参数：SaveDict/DeleteDict 需在事务内联动字典项（整存整取/级联删）
	DictModel interface {
		// Insert 新增字典（session 非空时在事务内执行），返回自增 id
		Insert(ctx context.Context, session sqlx.Session, data *Dict) (int64, error)
		// InsertOnDuplicate 按 uk_beid_code 冲突覆盖更新字典基础字段（session 非空时在事务内执行），
		// 返回字典 id（整存整取语义：编码已存在时等价于更新）
		InsertOnDuplicate(ctx context.Context, session sqlx.Session, data *Dict) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Dict, error)
		// FindOneByCode 按 beid + 字典编码查询（数据元值域校验用）
		FindOneByCode(ctx context.Context, session sqlx.Session, beid int64, dictCode string) (*Dict, error)
		// Update 更新字典基础字段（按 id+beid 定位，不改状态；session 非空时在事务内执行）
		Update(ctx context.Context, session sqlx.Session, data *Dict) error
		// Delete 删除字典（按 id+beid 定位；session 非空时在事务内执行）
		Delete(ctx context.Context, session sqlx.Session, id, beid int64) error
		// FindList 条件分页查询：beid 必带，keyword 对 dict_code/dict_name 模糊匹配；返回列表与总数
		FindList(ctx context.Context, beid int64, keyword string, page, size int64) ([]*Dict, int64, error)
	}

	defaultDictModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewDictModel 创建字典 model
func NewDictModel(conn sqlx.SqlConn) DictModel {
	return &defaultDictModel{
		conn:  conn,
		table: dictTable,
	}
}

// Insert 新增字典
func (m *defaultDictModel) Insert(ctx context.Context, session sqlx.Session, data *Dict) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, dictInsertColumns, placeholders(dictInsertFieldSize))
	ret, err := execCtx(ctx, m.conn, session, query,
		data.DictCode, data.DictName, data.Comment, data.Status, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// InsertOnDuplicate 按 uk_beid_code 冲突覆盖更新（整存整取幂等写入）
func (m *defaultDictModel) InsertOnDuplicate(ctx context.Context, session sqlx.Session, data *Dict) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s) "+
		"on duplicate key update `dict_name` = values(`dict_name`), `comment` = values(`comment`)",
		m.table, dictInsertColumns, placeholders(dictInsertFieldSize))
	ret, err := execCtx(ctx, m.conn, session, query,
		data.DictCode, data.DictName, data.Comment, data.Status, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询字典
func (m *defaultDictModel) FindOne(ctx context.Context, id, beid int64) (*Dict, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", dictRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Dict
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

// FindOneByCode 按 beid + 字典编码查询（session 非空时走事务连接：事务内 upsert 后回查必须读未提交数据）
func (m *defaultDictModel) FindOneByCode(ctx context.Context, session sqlx.Session, beid int64, dictCode string) (*Dict, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `dict_code` = ? limit 1", dictRows, m.table)
	var resp Dict
	err := queryRowCtx(ctx, m.conn, session, &resp, query, beid, dictCode)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 更新字典基础字段（按 id+beid 定位；status 为运营字段不随整存整取覆盖）
func (m *defaultDictModel) Update(ctx context.Context, session sqlx.Session, data *Dict) error {
	query := fmt.Sprintf("update %s set `dict_code` = ?, `dict_name` = ?, `comment` = ? "+
		"where `id` = ? and `beid` = ?", m.table)
	_, err := execCtx(ctx, m.conn, session, query,
		data.DictCode, data.DictName, data.Comment, data.Id, data.Beid)
	return err
}

// Delete 删除字典
func (m *defaultDictModel) Delete(ctx context.Context, session sqlx.Session, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := execCtx(ctx, m.conn, session, query, id, beid)
	return err
}

// FindList 字典列表（beid 多租户必带 + keyword 对 dict_code/dict_name 模糊匹配），返回列表与总数
func (m *defaultDictModel) FindList(ctx context.Context, beid int64, keyword string,
	page, size int64) ([]*Dict, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and (`dict_code` like ? or `dict_name` like ?)"
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
		dictRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Dict
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
