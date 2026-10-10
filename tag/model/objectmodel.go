package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	// 注意坑：builder.RawFieldNames 返回的是带反引号的列名（如 "`id`"），
	// stringx.Remove 剔除自增/时间列时必须用 "`id`" 形式，裸写 "id" 剔不掉
	objectFieldNames          = builder.RawFieldNames(&TagObject{})
	objectRows                = strings.Join(objectFieldNames, ",")
	objectRowsExpectAutoSet   = strings.Join(stringx.Remove(objectFieldNames, "`id`", "`created_at`"), ",")
	objectRowsWithPlaceHolder = strings.Join(stringx.Remove(objectFieldNames, "`id`", "`created_at`"), "=?,") + "=?"
)

const (
	objectTable           = "`dc_tag_object`"
	objectInsertFieldSize = 7 // 插入字段数（与 objectRowsExpectAutoSet 对应）
	objectDefaultPageSize = 10
	objectMaxPageSize     = 200
)

type (
	// TagObject dc_tag_object 行（标签对象：打标目标业务主表）
	TagObject struct {
		Id         int64     `db:"id"`
		ObjectType string    `db:"object_type"` //user/member/goods
		Name       string    `db:"name"`
		DsId       int64     `db:"ds_id"`      //对象主数据所在数据源
		DbName     string    `db:"db_name"`    //对象库
		TableName  string    `db:"table_name"` //对象主表
		PkColumn   string    `db:"pk_column"`  //主键列
		Beid       int64     `db:"beid"`
		CreatedAt  time.Time `db:"created_at"`
	}

	ObjectModel interface {
		// Insert 新增标签对象，返回自增 id
		Insert(ctx context.Context, data *TagObject) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*TagObject, error)
		// Update 全量更新（不含自增 id 与创建时间）
		Update(ctx context.Context, data *TagObject) error
		// Delete 按主键删除（仅删对象行，tag_def/tag_value 由业务侧另行处理）
		Delete(ctx context.Context, id, beid int64) error
		// FindList 条件分页（beid 必带，objectType 可选），返回列表与总数
		FindList(ctx context.Context, beid int64, objectType string, page, size int64) ([]*TagObject, int64, error)
	}

	defaultObjectModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewObjectModel 创建标签对象 model
func NewObjectModel(conn sqlx.SqlConn) ObjectModel {
	return &defaultObjectModel{
		conn:  conn,
		table: objectTable,
	}
}

// Insert 新增标签对象，created_at 由库默认值填充
func (m *defaultObjectModel) Insert(ctx context.Context, data *TagObject) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)", m.table, objectRowsExpectAutoSet, placeholders(objectInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query, data.ObjectType, data.Name, data.DsId, data.DbName,
		data.TableName, data.PkColumn, data.Beid)
	if err != nil {
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindOne 按主键查询标签对象（查询列 objectRows 9 列与 TagObject 9 字段严格对齐，
// 列数与字段数不一致时 go-zero 报 not matching destination to scan）
func (m *defaultObjectModel) FindOne(ctx context.Context, id, beid int64) (*TagObject, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", objectRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp TagObject
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

// Update 全量更新标签对象（带多租户过滤）
func (m *defaultObjectModel) Update(ctx context.Context, data *TagObject) error {
	query := fmt.Sprintf("update %s set %s where `id` = ? and `beid` = ?", m.table, objectRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.ObjectType, data.Name, data.DsId, data.DbName,
		data.TableName, data.PkColumn, data.Beid, data.Id, data.Beid)
	return err
}

// Delete 按主键删除标签对象
func (m *defaultObjectModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 条件分页查询标签对象，返回列表与总数
func (m *defaultObjectModel) FindList(ctx context.Context, beid int64, objectType string,
	page, size int64) ([]*TagObject, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	where := "`beid` = ?"
	args := []any{beid}
	if objectType != "" {
		where += " and `object_type` = ?"
		args = append(args, objectType)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页参数防御
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = objectDefaultPageSize
	}
	if size > objectMaxPageSize {
		size = objectMaxPageSize
	}
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?", objectRows, m.table, where)
	args = append(args, size, (page-1)*size)
	var list []*TagObject
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
