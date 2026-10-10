package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	archiveLogTable = "`dc_ops_archive_log`"
	// archiveLogInsertColumns 插入字段（不含自增 id），run_at 显式传入保证与任务回填时间一致
	archiveLogInsertColumns = "task_id,archived_rows,deleted_rows,status,msg,run_at"
	// archiveLogRows 查询列（msg 允许 NULL 用 ifnull 兜底；run_at 转毫秒时间戳并以 0 兜底，
	// 查询列数与 ArchiveLog 字段数一一对应，规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	archiveLogRows = "id,task_id,archived_rows,deleted_rows,status,ifnull(msg,'') as msg," +
		"ifnull(cast(unix_timestamp(run_at) * 1000 as signed), 0) as run_at"
)

// 归档执行日志（dc_ops_archive_log）
type (
	// ArchiveLog 归档执行日志行
	ArchiveLog struct {
		Id           int64  `db:"id"`
		TaskId       int64  `db:"task_id"`
		ArchivedRows int64  `db:"archived_rows"` // 本轮累计归档行数（含失败前已提交批）
		DeletedRows  int64  `db:"deleted_rows"`  // 本轮累计删除源表行数
		Status       string `db:"status"`        // success/failed
		Msg          string `db:"msg"`           // 失败原因与失败批号
		RunAt        int64  `db:"run_at"`        // 执行时间（毫秒时间戳）
	}

	// ArchiveLogModel 归档执行日志表操作接口
	ArchiveLogModel interface {
		// Insert 写入一次归档执行日志
		Insert(ctx context.Context, data *ArchiveLog) error
		// FindList 分页查询执行日志（taskId>0 时按任务过滤）；返回列表与总数
		FindList(ctx context.Context, taskId, page, size int64) ([]*ArchiveLog, int64, error)
	}

	defaultArchiveLogModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewArchiveLogModel 创建归档执行日志 model
func NewArchiveLogModel(conn sqlx.SqlConn) ArchiveLogModel {
	return &defaultArchiveLogModel{
		conn:  conn,
		table: archiveLogTable,
	}
}

// Insert 写入一次归档执行日志
func (m *defaultArchiveLogModel) Insert(ctx context.Context, data *ArchiveLog) error {
	query := fmt.Sprintf("insert into %s (%s) values (?,?,?,?,?,?)",
		m.table, archiveLogInsertColumns)
	_, err := m.conn.ExecCtx(ctx, query,
		data.TaskId, data.ArchivedRows, data.DeletedRows, data.Status, data.Msg,
		time.UnixMilli(data.RunAt))
	return err
}

// FindList 执行日志列表（taskId 可选过滤，按 id 倒序），返回列表与总数
func (m *defaultArchiveLogModel) FindList(ctx context.Context, taskId, page, size int64) ([]*ArchiveLog, int64, error) {
	page, size = normalizePage(page, size)

	where := "1 = 1"
	args := []any{}
	if taskId > 0 {
		where = "`task_id` = ?"
		args = append(args, taskId)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		archiveLogRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*ArchiveLog
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
