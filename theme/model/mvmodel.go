package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const mvTable = "`dc_theme_mv`"

// mvRows 查询列（显式 10 列，与 MvRow 10 字段严格对齐；last_status 可空列 ifnull 兜底，
// last_refresh_at 为可空 datetime——NULL 时 UNIX_TIMESTAMP 返回 NULL，一并 ifnull 兜底，
// 且 unix 秒乘 1000 转毫秒后 CAST 为 SIGNED，避免 decimal 秒值扫 int64 失败）
const mvRows = "`id`,`name`,`datasource_id`,`db_name`,`view_name`,`select_sql`,`refresh_mode`," +
	"IFNULL(`last_status`,'') AS `last_status`," +
	"CAST(IFNULL(UNIX_TIMESTAMP(`last_refresh_at`),0)*1000 AS SIGNED) AS `last_refresh_at`,`beid`"

// mvInsertCols 插入列清单（不含自增 id 与 created_at；last_status/last_refresh_at
// 由刷新动作 UpdateRunStatus 维护，建卡时留库默认值）
const mvInsertCols = "`name`,`datasource_id`,`db_name`,`view_name`,`select_sql`,`refresh_mode`"

// mvInsertFieldSize 插入字段数（与 mvInsertCols 严格对应）
const mvInsertFieldSize = 6

type (
	// MvRow 物化视图查询行：与 mvRows 的 10 列一一对应（不含时间列）
	MvRow struct {
		Id            int64  `db:"id"`
		Name          string `db:"name"`
		DatasourceId  int64  `db:"datasource_id"`
		DbName        string `db:"db_name"`
		ViewName      string `db:"view_name"`
		SelectSql     string `db:"select_sql"`
		RefreshMode   string `db:"refresh_mode"`
		LastStatus    string `db:"last_status"`
		LastRefreshAt int64  `db:"last_refresh_at"` //毫秒时间戳（未刷新过为 0）
		Beid          int64  `db:"beid"`
	}

	// ThemeMvModel 物化视图（dc_theme_mv）
	ThemeMvModel interface {
		// Insert 新增物化视图定义，返回自增 id
		Insert(ctx context.Context, data *MvRow) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*MvRow, error)
		// Update 全量更新定义（不含状态与时间列，带多租户过滤）
		Update(ctx context.Context, data *MvRow) error
		// Delete 按主键删除
		Delete(ctx context.Context, id, beid int64) error
		// FindList 条件分页（beid 必带）
		FindList(ctx context.Context, beid, page, size int64) ([]*MvRow, int64, error)
		// UpdateRunStatus 回写刷新结果：last_status + last_refresh_at（成功/失败均留痕）
		UpdateRunStatus(ctx context.Context, id, beid int64, status string) error
	}

	defaultThemeMvModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewThemeMvModel 创建物化视图 model
func NewThemeMvModel(conn sqlx.SqlConn) ThemeMvModel {
	return &defaultThemeMvModel{
		conn:  conn,
		table: mvTable,
	}
}

// Insert 新增物化视图定义，created_at/last_status/last_refresh_at 由库默认值填充
func (m *defaultThemeMvModel) Insert(ctx context.Context, data *MvRow) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)", m.table, mvInsertCols, placeholders(mvInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query, data.Name, data.DatasourceId, data.DbName,
		data.ViewName, data.SelectSql, data.RefreshMode)
	if err != nil {
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindOne 按主键查询物化视图（beid>0 时追加多租户过滤）
func (m *defaultThemeMvModel) FindOne(ctx context.Context, id, beid int64) (*MvRow, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", mvRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp MvRow
	if err := queryOne(ctx, m.conn, &resp, query+" limit 1", args...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update 全量更新物化视图定义（状态与刷新时间由 UpdateRunStatus 单独维护）
func (m *defaultThemeMvModel) Update(ctx context.Context, data *MvRow) error {
	query := fmt.Sprintf("update %s set `name` = ?, `datasource_id` = ?, `db_name` = ?, `view_name` = ?, "+
		"`select_sql` = ?, `refresh_mode` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.Name, data.DatasourceId, data.DbName,
		data.ViewName, data.SelectSql, data.RefreshMode, data.Id, data.Beid)
	return err
}

// Delete 按主键删除物化视图
func (m *defaultThemeMvModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 条件分页查询物化视图，返回列表与总数
func (m *defaultThemeMvModel) FindList(ctx context.Context, beid, page, size int64) ([]*MvRow, int64, error) {
	if err := ensureBeid(beid); err != nil {
		return nil, 0, err
	}
	where := "`beid` = ?"
	args := []any{beid}

	//总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	//分页
	page, size = normalizePage(page, size)
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?", mvRows, m.table, where)
	args = append(args, size, (page-1)*size)
	var list []*MvRow
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UpdateRunStatus 回写刷新结果：last_status=success/failed，last_refresh_at=本次动作时间
func (m *defaultThemeMvModel) UpdateRunStatus(ctx context.Context, id, beid int64, status string) error {
	query := fmt.Sprintf("update %s set `last_status` = ?, `last_refresh_at` = ? "+
		"where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, time.Now(), id, beid)
	return err
}
