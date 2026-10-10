package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	groupTable           = "`dc_tag_group`"
	groupDefaultPageSize = 10
	groupMaxPageSize     = 200
)

// groupRows 查询列（9 列与 Group 9 字段严格对齐，列数与字段数不一致时
// go-zero 扫宽 struct 报 not matching destination to scan）：
// circle_rule_sql 为 text 可空，用 ifnull 兜底避免扫描报错；
// last_run_at 可空，转秒级时间戳返回（未跑过为 0），pb 侧直接可用
const groupRows = "`id`,`name`,`object_id`,ifnull(`circle_rule_sql`,'') as `circle_rule_sql`,`est_count`,`materialized_table`,ifnull(unix_timestamp(`last_run_at`),0) as `last_run_at`,`beid`,`created_at`"

type (
	// Group dc_tag_group 行（标签人群/圈群定义）
	Group struct {
		Id                int64     `db:"id"`
		Name              string    `db:"name"`
		ObjectId          int64     `db:"object_id"`
		CircleRuleSql     string    `db:"circle_rule_sql"`    //圈选 SQL: select <pk> as object_id from ...
		EstCount          int64     `db:"est_count"`          //最近一次圈选成员数
		MaterializedTable string    `db:"materialized_table"` //人群物化表名（RunCircle 物化后回填）
		LastRunAt         int64     `db:"last_run_at"`        //最近一次圈选时间（秒级时间戳，0=未跑过）
		Beid              int64     `db:"beid"`
		CreatedAt         time.Time `db:"created_at"`
	}

	GroupModel interface {
		// Insert 新增人群定义，返回自增 id
		Insert(ctx context.Context, data *Group) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Group, error)
		// FindList 分页列表（beid 必带），返回列表与总数
		FindList(ctx context.Context, beid, page, size int64) ([]*Group, int64, error)
		// UpdateRunStatus 圈选完成后回写运行状态：成员数/物化表名，last_run_at 落 now(3)
		UpdateRunStatus(ctx context.Context, id, beid, estCount int64, materializedTable string) error
		// Delete 删除人群定义（仅删定义行，人群物化表留档不删）
		Delete(ctx context.Context, id, beid int64) error
	}

	defaultGroupModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewGroupModel 创建标签人群 model
func NewGroupModel(conn sqlx.SqlConn) GroupModel {
	return &defaultGroupModel{
		conn:  conn,
		table: groupTable,
	}
}

// Insert 新增人群定义：est_count/materialized_table/last_run_at 首次圈选时由
// RunCircle 回填，created_at 由库默认值填充
func (m *defaultGroupModel) Insert(ctx context.Context, data *Group) (int64, error) {
	query := fmt.Sprintf("insert into %s (`name`,`object_id`,`circle_rule_sql`,`beid`) values (?,?,?,?)", m.table)
	ret, err := m.conn.ExecCtx(ctx, query, data.Name, data.ObjectId, data.CircleRuleSql, data.Beid)
	if err != nil {
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindOne 按主键查询人群定义
func (m *defaultGroupModel) FindOne(ctx context.Context, id, beid int64) (*Group, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", groupRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Group
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

// FindList 分页查询人群定义（按 id 倒序）
func (m *defaultGroupModel) FindList(ctx context.Context, beid, page, size int64) ([]*Group, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where `beid` = ?", m.table)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, beid); err != nil {
		return nil, 0, err
	}

	// 分页参数防御
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = groupDefaultPageSize
	}
	if size > groupMaxPageSize {
		size = groupMaxPageSize
	}
	listQuery := fmt.Sprintf("select %s from %s where `beid` = ? order by `id` desc limit ? offset ?", groupRows, m.table)
	var list []*Group
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, beid, size, (page-1)*size); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UpdateRunStatus 回写圈选运行状态：成员数/物化表名/最近圈选时间（now(3)），
// 由 RunCircle 物化成功后调用（此时物化表名已确定为 dc_tag_group_member_<group_id>）
func (m *defaultGroupModel) UpdateRunStatus(ctx context.Context, id, beid, estCount int64, materializedTable string) error {
	query := fmt.Sprintf("update %s set `est_count` = ?, `materialized_table` = ?, `last_run_at` = now(3) where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, estCount, materializedTable, id, beid)
	return err
}

// Delete 删除人群定义（仅删定义行；人群物化表留档，由运维侧另行清理）
func (m *defaultGroupModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}
