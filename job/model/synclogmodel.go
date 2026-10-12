package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	syncLogTable = "`dc_sync_task_log`"
	// syncLogInsertColumns 插入字段（不含自增 id），run_at 显式传入保证与任务状态回填时间一致
	syncLogInsertColumns = "task_id,read_rows,write_rows,status,msg,run_at"
	// syncLogRows 查询列（msg 允许 NULL 用 ifnull 兜底；run_at 转毫秒时间戳并以 0 兜底，
	// 查询列数与 SyncLog 字段数一一对应，规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	syncLogRows = "id,task_id,read_rows,write_rows,status,ifnull(msg,'') as msg," +
		"ifnull(cast(unix_timestamp(run_at) * 1000 as signed), 0) as run_at"
)

// EnsureSyncLogTableDDL dc_sync_task_log 幂等建表 DDL（svc 启动时自动执行，无需手工迁移）。
// 一次内置同步执行落一行：成功/失败都落，失败 msg 记失败批号与原因（与 dc_ops_archive_log 同风格）
const EnsureSyncLogTableDDL = "create table if not exists " + syncLogTable + " (" +
	"`id` bigint unsigned NOT NULL AUTO_INCREMENT," +
	"`task_id` bigint NOT NULL COMMENT '任务id(dc_sync_task.id)'," +
	"`read_rows` bigint NOT NULL DEFAULT 0 COMMENT '累计读取行数(含失败前已读批)'," +
	"`write_rows` bigint NOT NULL DEFAULT 0 COMMENT '累计写入行数(已提交批)'," +
	"`status` varchar(16) NOT NULL DEFAULT 'success' COMMENT 'success/failed'," +
	"`msg` varchar(1024) NOT NULL DEFAULT '' COMMENT '失败原因(含失败批号)'," +
	"`run_at` datetime(3) DEFAULT NULL COMMENT '执行时间'," +
	"PRIMARY KEY (`id`) USING BTREE," +
	"KEY `idx_task_run` (`task_id`,`run_at`)" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内置同步执行日志表'"

// EnsureSyncLogTable svc 启动时自动建日志表（幂等 DDL；新建环境免迁移，
// 运行库缺表仅影响执行日志落库与查询，不影响同步主流程）
func EnsureSyncLogTable(conn sqlx.SqlConn) error {
	if _, err := conn.ExecCtx(context.Background(), EnsureSyncLogTableDDL); err != nil {
		return fmt.Errorf("建表 %s 失败: %w", syncLogTable, err)
	}
	return nil
}

// 内置同步执行日志（dc_sync_task_log）
type (
	// SyncLog 内置同步执行日志行
	SyncLog struct {
		Id        int64  `db:"id"`
		TaskId    int64  `db:"task_id"`
		ReadRows  int64  `db:"read_rows"`  // 累计读取行数（含失败前已读批）
		WriteRows int64  `db:"write_rows"` // 累计写入行数（失败中断时为已提交批）
		Status    string `db:"status"`     // success/failed
		Msg       string `db:"msg"`        // 失败原因与失败批号
		RunAt     int64  `db:"run_at"`     // 执行时间（毫秒时间戳）
	}

	// SyncLogModel 内置同步执行日志表操作接口
	SyncLogModel interface {
		// Insert 写入一次内置同步执行日志
		Insert(ctx context.Context, data *SyncLog) error
		// FindList 分页查询执行日志（taskId>0 时按任务过滤）；返回列表与总数
		FindList(ctx context.Context, taskId, page, size int64) ([]*SyncLog, int64, error)
	}

	defaultSyncLogModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewSyncLogModel 创建内置同步执行日志 model
func NewSyncLogModel(conn sqlx.SqlConn) SyncLogModel {
	return &defaultSyncLogModel{
		conn:  conn,
		table: syncLogTable,
	}
}

// Insert 写入一次内置同步执行日志
func (m *defaultSyncLogModel) Insert(ctx context.Context, data *SyncLog) error {
	query := fmt.Sprintf("insert into %s (%s) values (?,?,?,?,?,?)",
		m.table, syncLogInsertColumns)
	_, err := m.conn.ExecCtx(ctx, query,
		data.TaskId, data.ReadRows, data.WriteRows, data.Status, data.Msg,
		time.UnixMilli(data.RunAt))
	return err
}

// FindList 执行日志列表（taskId 可选过滤，按 id 倒序），返回列表与总数
func (m *defaultSyncLogModel) FindList(ctx context.Context, taskId, page, size int64) ([]*SyncLog, int64, error) {
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
		syncLogRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*SyncLog
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
