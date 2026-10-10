package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	jobTable           = "`dc_quality_job`"
	jobInsertFieldSize = 6 // 插入字段数（与 jobInsertColumns 对应）
	// jobInsertColumns 插入字段（不含自增 id、时间列与运行状态列，last_status/last_run_at 由 RunJob 回填）
	jobInsertColumns = "name,datasource_id,db_name,rule_ids,cron,beid"
	// jobRows 查询列（last_run_at 允许 NULL，转毫秒时间戳并以 0 兜底避免扫描报错；
	// 查询列数与 Job 字段数一一对应，规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	jobRows = "id,name,datasource_id,db_name,rule_ids,cron,last_status," +
		"ifnull(cast(unix_timestamp(last_run_at) * 1000 as signed), 0) as last_run_at," +
		"beid,created_at,updated_at"
)

// 质量任务（dc_quality_job）
type (
	// Job 质量任务行
	Job struct {
		Id           int64     `db:"id"`
		Name         string    `db:"name"`
		DatasourceId int64     `db:"datasource_id"` // 目标数据源
		DbName       string    `db:"db_name"`       // 目标库（空则使用数据源默认库）
		RuleIds      string    `db:"rule_ids"`      // 规则id，逗号分隔
		Cron         string    `db:"cron"`          // cron 表达式（Phase 3 调度接入，本期仅存档）
		LastStatus   string    `db:"last_status"`   // success/failed/partial
		LastRunAt    int64     `db:"last_run_at"`   // 最近执行时间（毫秒时间戳，未运行为 0）
		Beid         int64     `db:"beid"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}

	// JobModel 质量任务表操作接口
	JobModel interface {
		// Insert 新增任务，返回自增 id
		Insert(ctx context.Context, data *Job) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Job, error)
		// Update 更新任务可编辑字段（按 id+beid 定位）
		Update(ctx context.Context, data *Job) error
		// Delete 删除任务（按 id+beid 定位）
		Delete(ctx context.Context, id, beid int64) error
		// FindList 任务分页列表：beid 必带；返回列表与总数
		FindList(ctx context.Context, beid, page, size int64) ([]*Job, int64, error)
		// UpdateRunStatus 任务执行完成后回填最近状态与执行时间
		UpdateRunStatus(ctx context.Context, id, beid int64, status string, runAt time.Time) error
	}

	defaultJobModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewJobModel 创建质量任务 model
func NewJobModel(conn sqlx.SqlConn) JobModel {
	return &defaultJobModel{
		conn:  conn,
		table: jobTable,
	}
}

// Insert 新增质量任务
func (m *defaultJobModel) Insert(ctx context.Context, data *Job) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, jobInsertColumns, placeholders(jobInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Name, data.DatasourceId, data.DbName, data.RuleIds, data.Cron, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询任务
func (m *defaultJobModel) FindOne(ctx context.Context, id, beid int64) (*Job, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", jobRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Job
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

// Update 更新任务（按 id+beid 定位，运行状态列由 UpdateRunStatus 维护）
func (m *defaultJobModel) Update(ctx context.Context, data *Job) error {
	query := fmt.Sprintf("update %s set `name` = ?, `datasource_id` = ?, `db_name` = ?, "+
		"`rule_ids` = ?, `cron` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Name, data.DatasourceId, data.DbName, data.RuleIds, data.Cron,
		data.Id, data.Beid)
	return err
}

// Delete 删除任务（结果历史不级联清理，供问题追溯）
func (m *defaultJobModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 任务列表（beid 多租户必带），返回列表与总数
func (m *defaultJobModel) FindList(ctx context.Context, beid, page, size int64) ([]*Job, int64, error) {
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
		jobRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Job
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UpdateRunStatus 任务执行完成后回填 last_status/last_run_at
func (m *defaultJobModel) UpdateRunStatus(ctx context.Context, id, beid int64, status string, runAt time.Time) error {
	query := fmt.Sprintf("update %s set `last_status` = ?, `last_run_at` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, runAt, id, beid)
	return err
}
