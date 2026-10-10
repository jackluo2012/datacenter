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
	scriptTable = "`dc_ide_script`"
	// scriptInsertColumns 插入字段（不含自增 id 与时间列；version 新建固定 1，更新时在库上自增）
	scriptInsertColumns = "`name`,`sql_text`,`datasource_id`,`db_name`,`beid`"
)

// scriptRows 查询列（last_run_at 允许 NULL，转毫秒时间戳并以 0 兜底避免扫描报错；
// 列数与 Script 字段数一一对应（10 列对 10 字段），规避 go-zero 少列查询扫宽 struct
// 的 "not matching destination to scan" 报错）
const scriptRows = "`id`,`name`,`sql_text`,`version`,`datasource_id`,`db_name`," +
	"ifnull(cast(unix_timestamp(`last_run_at`) * 1000 as signed), 0) as `last_run_at`," +
	"`beid`,`created_at`,`updated_at`"

// Script dc_ide_script 行（SQL 脚本）
type (
	Script struct {
		Id           int64     `db:"id"`
		Name         string    `db:"name"`
		SqlText      string    `db:"sql_text"` // SQL 内容（mediumtext）
		Version      int32     `db:"version"`  // 版本号，每次保存 +1
		DatasourceId int64     `db:"datasource_id"`
		DbName       string    `db:"db_name"`     // 目标库（空则使用数据源默认库）
		LastRunAt    int64     `db:"last_run_at"` // 最近执行时间（毫秒时间戳，未运行为 0）
		Beid         int64     `db:"beid"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}

	// ScriptModel SQL 脚本表操作接口
	ScriptModel interface {
		// Save 保存脚本：id=0 新建（version=1）；id>0 更新内容且 version=version+1（回读整行返回新版本号）
		Save(ctx context.Context, data *Script) (*Script, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Script, error)
		// FindList 分页列表（beid 必带，keyword 模糊匹配 name），返回列表与总数
		FindList(ctx context.Context, beid int64, keyword string, page, size int64) ([]*Script, int64, error)
		// Delete 删除脚本（按 id+beid 定位）
		Delete(ctx context.Context, id, beid int64) error
		// UpdateLastRunAt 脚本执行完成后回填最近执行时间
		UpdateLastRunAt(ctx context.Context, id, beid int64, runAt time.Time) error
	}

	defaultScriptModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewScriptModel 创建 SQL 脚本 model
func NewScriptModel(conn sqlx.SqlConn) ScriptModel {
	return &defaultScriptModel{
		conn:  conn,
		table: scriptTable,
	}
}

// Save 保存脚本（新建与更新主键不同，insert ... on duplicate 不适用，拆成两条链路）：
// id=0 走 insert，version 固定 1；id>0 走 update，version=version+1 在库上自增
// （规避读-改-写的并发丢版本），随后回读整行（含新版本号）返回
func (m *defaultScriptModel) Save(ctx context.Context, data *Script) (*Script, error) {
	if data.Id <= 0 {
		query := fmt.Sprintf("insert into %s (%s,`version`) values (?,?,?,?,?,1)", m.table, scriptInsertColumns)
		ret, err := m.conn.ExecCtx(ctx, query,
			data.Name, data.SqlText, data.DatasourceId, data.DbName, data.Beid)
		if err != nil {
			return nil, err
		}
		id, err := ret.LastInsertId()
		if err != nil {
			return nil, err
		}
		data.Id = id
		data.Version = 1
		return data, nil
	}

	// 更新：version 在库上 +1，内容列整体覆盖
	query := fmt.Sprintf("update %s set `name` = ?, `sql_text` = ?, `datasource_id` = ?, `db_name` = ?, "+
		"`version` = `version` + 1 where `id` = ? and `beid` = ?", m.table)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Name, data.SqlText, data.DatasourceId, data.DbName, data.Id, data.Beid)
	if err != nil {
		return nil, err
	}
	// 影响行数为 0：记录不存在或不归属当前租户（并发被删的兜底）
	if affected, aerr := ret.RowsAffected(); aerr == nil && affected == 0 {
		return nil, ErrNotFound
	}
	// 回读更新后的整行（含新版本号）
	return m.FindOne(ctx, data.Id, data.Beid)
}

// FindOne 按主键查询脚本
func (m *defaultScriptModel) FindOne(ctx context.Context, id, beid int64) (*Script, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", scriptRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Script
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

// FindList 分页查询脚本列表（beid 多租户必带），keyword 非空时模糊匹配 name
func (m *defaultScriptModel) FindList(ctx context.Context, beid int64, keyword string, page, size int64) ([]*Script, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and `name` like ?"
		args = append(args, "%"+escapeLike(kw)+"%")
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		scriptRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Script
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Delete 删除脚本（按 id+beid 定位，防止越权删除其他租户脚本）
func (m *defaultScriptModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// UpdateLastRunAt 脚本执行完成后回填最近执行时间
func (m *defaultScriptModel) UpdateLastRunAt(ctx context.Context, id, beid int64, runAt time.Time) error {
	query := fmt.Sprintf("update %s set `last_run_at` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, runAt, id, beid)
	return err
}
