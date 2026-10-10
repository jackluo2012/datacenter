package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	grantTable = "`dc_sec_grant`"
	// grantInsertColumns 插入字段（不含自增 id，created_at 由数据库默认当前时间）
	grantInsertColumns = "`subject`,`resource`,`resource_id`,`level_code`,`expire_at`,`beid`"
	// grantRows 查询列（expire_at 为可空 datetime 列，扫入 sql.NullTime，NULL 表示永久；
	// 7 列与 Grant 7 字段一一对应，规避扫描错位）
	grantRows = "`id`,`subject`,`resource`,`resource_id`,`level_code`,`expire_at`,`beid`"
)

// 授权矩阵（dc_sec_grant）：主体 → 资源 → 许可的最高密级
type (
	// Grant 授权行
	Grant struct {
		Id         int64        `db:"id"`
		Subject    string       `db:"subject"`  //user:1 / app:app_key / role:admin
		Resource   string       `db:"resource"` //datasource/api/table/asset
		ResourceId string       `db:"resource_id"`
		LevelCode  string       `db:"level_code"` //许可的最高密级
		ExpireAt   sql.NullTime `db:"expire_at"`  //NULL 表示永久
		Beid       int64        `db:"beid"`
	}

	// GrantModel 授权矩阵表操作接口
	GrantModel interface {
		// Save 按 uk_sub_res upsert（重复授权覆盖许可密级与过期时间），返回记录 id
		Save(ctx context.Context, data *Grant) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Grant, error)
		// FindBySubRes 按主体 + 资源定位授权（CheckPerm 鉴权入口）
		FindBySubRes(ctx context.Context, beid int64, subject, resource, resourceId string) (*Grant, error)
		// Delete 删除授权（按 id+beid 定位）
		Delete(ctx context.Context, id, beid int64) error
		// List 条件分页：beid 必带，subject/resource 非空时精确过滤；返回列表与总数
		List(ctx context.Context, beid int64, subject, resource string, page, size int64) ([]*Grant, int64, error)
	}

	defaultGrantModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewGrantModel 创建授权矩阵 model
func NewGrantModel(conn sqlx.SqlConn) GrantModel {
	return &defaultGrantModel{
		conn:  conn,
		table: grantTable,
	}
}

// Save 按 uk_sub_res（subject+resource+resource_id）upsert 授权。
// 更新分支借 `id = last_insert_id(id)` 让 LastInsertId 返回既有主键（insert/update 两分支都可靠）；
// expire_at 以 sql.NullTime 落库：Valid=false 时写 NULL（永久）
func (m *defaultGrantModel) Save(ctx context.Context, data *Grant) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?) "+
		"on duplicate key update `id` = last_insert_id(`id`), `level_code` = values(`level_code`), "+
		"`expire_at` = values(`expire_at`), `beid` = values(`beid`)",
		m.table, grantInsertColumns)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Subject, data.Resource, data.ResourceId, data.LevelCode, data.ExpireAt, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询授权
func (m *defaultGrantModel) FindOne(ctx context.Context, id, beid int64) (*Grant, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", grantRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Grant
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

// FindBySubRes 按主体 + 资源定位授权（uk_sub_res 精确命中）
func (m *defaultGrantModel) FindBySubRes(ctx context.Context, beid int64,
	subject, resource, resourceId string) (*Grant, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where `subject` = ? and `resource` = ? "+
		"and `resource_id` = ? and `beid` = ? limit 1", grantRows, m.table)
	var resp Grant
	err := m.conn.QueryRowCtx(ctx, &resp, query, subject, resource, resourceId, beid)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Delete 删除授权
func (m *defaultGrantModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// List 授权列表（beid 多租户必带，subject/resource 非空时精确过滤），返回列表与总数
func (m *defaultGrantModel) List(ctx context.Context, beid int64,
	subject, resource string, page, size int64) ([]*Grant, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if subject != "" {
		where += " and `subject` = ?"
		args = append(args, subject)
	}
	if resource != "" {
		where += " and `resource` = ?"
		args = append(args, resource)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		grantRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Grant
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
