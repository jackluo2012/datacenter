package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	archiveTaskTable           = "`dc_ops_archive_task`"
	archiveTaskInsertFieldSize = 10 // 插入字段数（与 archiveTaskInsertColumns 对应）
	// archiveTaskInsertColumns 插入字段（不含自增 id、时间列与运行状态列，last_status/last_run_at 由 RunArchive 回填）
	archiveTaskInsertColumns = "name,datasource_id,db_name,table_name,condition_sql,archive_table,delete_source,batch_size,cron,beid"
	// archiveTaskRows 查询列（db_name/condition_sql/cron/last_status 允许 NULL 用 ifnull 兜底；
	// last_run_at 转毫秒时间戳并以 0 兜底，查询列数与 ArchiveTask 字段数一一对应，
	// 规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	archiveTaskRows = "id,name,datasource_id,ifnull(db_name,'') as db_name,table_name," +
		"ifnull(condition_sql,'') as condition_sql,archive_table,delete_source,batch_size," +
		"ifnull(cron,'') as cron,ifnull(last_status,'') as last_status," +
		"ifnull(cast(unix_timestamp(last_run_at) * 1000 as signed), 0) as last_run_at," +
		"beid,created_at,updated_at"
)

// 归档任务（dc_ops_archive_task）
type (
	// ArchiveTask 归档任务行
	ArchiveTask struct {
		Id           int64     `db:"id"`
		Name         string    `db:"name"`
		DatasourceId int64     `db:"datasource_id"` // 源数据源
		DbName       string    `db:"db_name"`       // 源库（空则使用数据源默认库）
		TableName    string    `db:"table_name"`    // 源表
		ConditionSql string    `db:"condition_sql"` // 归档条件（where 片段，白名单校验，可空=整表归档）
		ArchiveTable string    `db:"archive_table"` // 归档表
		DeleteSource int64     `db:"delete_source"` // 1.归档后删源 0.仅插入
		BatchSize    int64     `db:"batch_size"`    // 分批大小
		Cron         string    `db:"cron"`          // cron 表达式（Phase 3 调度接入，本期仅存档）
		LastStatus   string    `db:"last_status"`   // success/failed
		LastRunAt    int64     `db:"last_run_at"`   // 最近执行时间（毫秒时间戳，未运行为 0）
		Beid         int64     `db:"beid"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}

	// ArchiveTaskModel 归档任务表操作接口
	ArchiveTaskModel interface {
		// Insert 新增任务，返回自增 id
		Insert(ctx context.Context, data *ArchiveTask) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*ArchiveTask, error)
		// Delete 删除任务（按 id+beid 定位；历史执行日志不级联清理，供问题追溯）
		Delete(ctx context.Context, id, beid int64) error
		// FindList 任务分页列表：beid 必带；返回列表与总数
		FindList(ctx context.Context, beid, page, size int64) ([]*ArchiveTask, int64, error)
		// UpdateRunStatus 归档执行完成后回填最近状态与执行时间
		UpdateRunStatus(ctx context.Context, id, beid int64, status string, runAt time.Time) error
	}

	defaultArchiveTaskModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewArchiveTaskModel 创建归档任务 model
func NewArchiveTaskModel(conn sqlx.SqlConn) ArchiveTaskModel {
	return &defaultArchiveTaskModel{
		conn:  conn,
		table: archiveTaskTable,
	}
}

// Insert 新增归档任务
func (m *defaultArchiveTaskModel) Insert(ctx context.Context, data *ArchiveTask) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, archiveTaskInsertColumns, placeholders(archiveTaskInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Name, data.DatasourceId, data.DbName, data.TableName, data.ConditionSql,
		data.ArchiveTable, data.DeleteSource, data.BatchSize, data.Cron, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询归档任务
func (m *defaultArchiveTaskModel) FindOne(ctx context.Context, id, beid int64) (*ArchiveTask, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", archiveTaskRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp ArchiveTask
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

// Delete 删除归档任务（执行日志不级联清理，供问题追溯）
func (m *defaultArchiveTaskModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 任务列表（beid 多租户必带），返回列表与总数
func (m *defaultArchiveTaskModel) FindList(ctx context.Context, beid, page, size int64) ([]*ArchiveTask, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		archiveTaskRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*ArchiveTask
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UpdateRunStatus 归档执行完成后回填 last_status/last_run_at
func (m *defaultArchiveTaskModel) UpdateRunStatus(ctx context.Context, id, beid int64, status string, runAt time.Time) error {
	query := fmt.Sprintf("update %s set `last_status` = ?, `last_run_at` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, runAt, id, beid)
	return err
}
