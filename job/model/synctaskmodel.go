package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	taskTable           = "`dc_sync_task`"
	taskInsertFieldSize = 7 // 插入字段数（与 taskInsertColumns 对应）
	// taskInsertColumns 插入字段（不含自增 id、时间列与部署/运行状态列，
	// ds_workflow_code/status 由部署回填，last_run_at 由触发运行回填）
	taskInsertColumns = "name,source_ds_id,target_ds_id,mode,mapping_json,cron,beid"
	// taskRows 查询列（mapping_json/last_run_at 允许 NULL，分别用 ifnull 兜底为空串/0 毫秒时间戳，
	// 避免扫描报错；查询列数与 SyncTask 字段数一一对应，
	// 规避 go-zero 少列扫宽 struct 的 not matching destination to scan 报错）
	taskRows = "id,name,source_ds_id,target_ds_id,mode," +
		"ifnull(mapping_json,'') as mapping_json,cron,ds_workflow_code,status," +
		"ifnull(cast(unix_timestamp(last_run_at) * 1000 as signed), 0) as last_run_at," +
		"beid,created_at,updated_at"
)

// 同步任务（管控面：本平台管理定义，执行面推送 DolphinScheduler/SeaTunnel）
type (
	// SyncTask dc_sync_task 行
	SyncTask struct {
		Id             int64     `db:"id"`
		Name           string    `db:"name"`
		SourceDsId     int64     `db:"source_ds_id"`
		TargetDsId     int64     `db:"target_ds_id"`
		Mode           string    `db:"mode"` // offline/cdc
		MappingJson    string    `db:"mapping_json"`
		Cron           string    `db:"cron"` // 调度周期（推送 DS 时使用，本期仅存档）
		DsWorkflowCode int64     `db:"ds_workflow_code"`
		Status         string    `db:"status"`
		LastRunAt      int64     `db:"last_run_at"` // 最近运行时间（毫秒时间戳，未运行为 0）
		Beid           int64     `db:"beid"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	// SyncTaskModel 同步任务表操作接口
	SyncTaskModel interface {
		// Insert 新增同步任务（草稿态），返回自增 id
		Insert(ctx context.Context, data *SyncTask) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*SyncTask, error)
		// FindList 条件分页查询：beid 必带，status 可选过滤；返回列表与总数
		FindList(ctx context.Context, beid int64, status string, page, size int64) ([]*SyncTask, int64, error)
		// UpdateDeploy 部署回填：写入 DS 工作流编码并把状态流转到 deployed（按 id+beid 定位）
		UpdateDeploy(ctx context.Context, id, beid, workflowCode int64, status string) error
		// UpdateStatus 状态流转（deployed/running/failed，按 id+beid 定位）
		UpdateStatus(ctx context.Context, id, beid int64, status string) error
		// UpdateRunStatus 触发运行后回填状态与最近运行时间（按 id+beid 定位）
		UpdateRunStatus(ctx context.Context, id, beid int64, status string, runAt time.Time) error
		// Delete 删除任务（按 id+beid 定位）
		Delete(ctx context.Context, id, beid int64) error
	}

	defaultSyncTaskModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewSyncTaskModel 创建同步任务 model
func NewSyncTaskModel(conn sqlx.SqlConn) SyncTaskModel {
	return &defaultSyncTaskModel{
		conn:  conn,
		table: taskTable,
	}
}

// Insert 新增同步任务（草稿态）
func (m *defaultSyncTaskModel) Insert(ctx context.Context, data *SyncTask) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, taskInsertColumns, placeholders(taskInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Name, data.SourceDsId, data.TargetDsId, data.Mode,
		data.MappingJson, data.Cron, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询同步任务
func (m *defaultSyncTaskModel) FindOne(ctx context.Context, id, beid int64) (*SyncTask, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", taskRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp SyncTask
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

// FindList 任务列表（beid 多租户必带 + status 可选过滤），返回列表与总数
func (m *defaultSyncTaskModel) FindList(ctx context.Context, beid int64, status string,
	page, size int64) ([]*SyncTask, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if status != "" {
		where += " and `status` = ?"
		args = append(args, status)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		taskRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*SyncTask
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UpdateDeploy 部署回填：写入 DS 工作流编码并把状态流转到 deployed
func (m *defaultSyncTaskModel) UpdateDeploy(ctx context.Context, id, beid, workflowCode int64, status string) error {
	query := fmt.Sprintf("update %s set `ds_workflow_code` = ?, `status` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, workflowCode, status, id, beid)
	return err
}

// UpdateStatus 状态流转
func (m *defaultSyncTaskModel) UpdateStatus(ctx context.Context, id, beid int64, status string) error {
	query := fmt.Sprintf("update %s set `status` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id, beid)
	return err
}

// UpdateRunStatus 触发运行后回填状态与最近运行时间
func (m *defaultSyncTaskModel) UpdateRunStatus(ctx context.Context, id, beid int64, status string, runAt time.Time) error {
	query := fmt.Sprintf("update %s set `status` = ?, `last_run_at` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, runAt, id, beid)
	return err
}

// Delete 删除任务
func (m *defaultSyncTaskModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}
