package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	metricRelTable = "`dc_metric_rel`"
)

// metricRelInsertColumns 插入列（不含自增 id）
const metricRelInsertColumns = "`derive_metric_id`,`atom_metric_id`,`beid`"

// 指标关系（派生指标 → 原子指标引用）
type (
	// MetricRel dc_metric_rel 行
	MetricRel struct {
		Id             int64 `db:"id"`
		DeriveMetricId int64 `db:"derive_metric_id"`
		AtomMetricId   int64 `db:"atom_metric_id"`
		Beid           int64 `db:"beid"`
	}

	MetricRelModel interface {
		// BindAll 覆盖式绑定派生指标的原子引用：事务内先删后插，幂等覆盖
		BindAll(ctx context.Context, session sqlx.Session, deriveMetricId, beid int64, atomIds []int64) error
		// DeleteByDeriveId 删除派生指标的全部引用（删除派生指标级联用）
		DeleteByDeriveId(ctx context.Context, session sqlx.Session, deriveMetricId int64) error
		// ListAtoms 查询派生指标引用的原子指标 id 列表
		ListAtoms(ctx context.Context, deriveMetricId int64) ([]int64, error)
		// CountByAtomId 统计原子指标被派生引用的次数（删除原子指标前置校验）
		CountByAtomId(ctx context.Context, atomMetricId int64) (int64, error)
	}

	defaultMetricRelModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewMetricRelModel 创建指标关系 model
func NewMetricRelModel(conn sqlx.SqlConn) MetricRelModel {
	return &defaultMetricRelModel{
		conn:  conn,
		table: metricRelTable,
	}
}

// BindAll 覆盖式绑定：事务内先删该派生指标的全部旧引用，再批量插入新引用（atomIds 自动去重）；
// uk_rel(derive_metric_id,atom_metric_id) 冲突时 insert ignore 兜底，保证重复调用幂等
func (m *defaultMetricRelModel) BindAll(ctx context.Context, session sqlx.Session,
	deriveMetricId, beid int64, atomIds []int64) error {
	if err := m.DeleteByDeriveId(ctx, session, deriveMetricId); err != nil {
		return err
	}
	uniq := UniqInt64s(atomIds)
	if len(uniq) == 0 {
		return nil
	}
	rows := strings.TrimSuffix(strings.Repeat("(?,?,?),", len(uniq)), ",")
	query := fmt.Sprintf("insert ignore into %s (%s) values %s", m.table, metricRelInsertColumns, rows)
	args := make([]any, 0, len(uniq)*3)
	for _, id := range uniq {
		args = append(args, deriveMetricId, id, beid)
	}
	if _, err := execCtx(ctx, m.conn, session, query, args...); err != nil {
		return fmt.Errorf("批量绑定原子指标失败: %w", err)
	}
	return nil
}

// DeleteByDeriveId 删除派生指标的全部原子引用
func (m *defaultMetricRelModel) DeleteByDeriveId(ctx context.Context, session sqlx.Session, deriveMetricId int64) error {
	query := fmt.Sprintf("delete from %s where `derive_metric_id` = ?", m.table)
	_, err := execCtx(ctx, m.conn, session, query, deriveMetricId)
	return err
}

// atomIdRow ListAtoms 专用扫描结构（与 select atom_metric_id 单列对齐：
// go-zero 对少列查询扫宽 struct 会报 not matching destination to scan）
type atomIdRow struct {
	AtomMetricId int64 `db:"atom_metric_id"`
}

// ListAtoms 查询派生指标引用的原子指标 id 列表
func (m *defaultMetricRelModel) ListAtoms(ctx context.Context, deriveMetricId int64) ([]int64, error) {
	query := fmt.Sprintf("select `atom_metric_id` from %s where `derive_metric_id` = ?", m.table)
	var rows []*atomIdRow
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, deriveMetricId); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.AtomMetricId)
	}
	return ids, nil
}

// CountByAtomId 统计原子指标被派生指标引用的次数（>0 表示被引用）
func (m *defaultMetricRelModel) CountByAtomId(ctx context.Context, atomMetricId int64) (int64, error) {
	query := fmt.Sprintf("select count(*) as cnt from %s where `atom_metric_id` = ?", m.table)
	var cnt int64
	if err := m.conn.QueryRowCtx(ctx, &cnt, query, atomMetricId); err != nil {
		return 0, err
	}
	return cnt, nil
}
