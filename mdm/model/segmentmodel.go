package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const segmentTable = "`dc_mdm_segment`"

// segmentRows 查询列（显式 7 列，与 segmentRow 7 字段严格对齐，不含 updated_at）
const segmentRows = "`id`,`model_id`,`prefix`,`date_pattern`,`current_value`,`step`,`beid`"

type (
	// SegmentRow 码段行：与 segmentRows 的 7 列一一对应
	SegmentRow struct {
		Id           int64  `db:"id"`
		ModelId      int64  `db:"model_id"`
		Prefix       string `db:"prefix"`
		DatePattern  string `db:"date_pattern"` //Go 时间格式（如 20060102），空为无日期段
		CurrentValue int64  `db:"current_value"`
		Step         int    `db:"step"`
		Beid         int64  `db:"beid"`
	}

	// SegmentModel 编码码段（dc_mdm_segment，号段模式）
	SegmentModel interface {
		// FindOrCreateByModel 按模型取码段配置：存在直接返回；
		// 不存在则按入参默认值创建（uk_model 唯一键，并发首取撞键时重查兜底）
		FindOrCreateByModel(ctx context.Context, modelId, beid int64, prefix, datePattern string, step int) (*SegmentRow, error)
		// NextValue 号段取号：事务内 update current_value=current_value+step 后回读，
		// 依赖 InnoDB 行锁串行化并发取号，返回推进后的码段
		NextValue(ctx context.Context, modelId int64) (*SegmentRow, error)
		// DeleteByModel 按模型删除码段（供 DeleteModel 事务内级联清理，session 为空时走普通连接）
		DeleteByModel(ctx context.Context, session sqlx.Session, modelId, beid int64) error
	}

	defaultSegmentModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewSegmentModel 创建码段 model
func NewSegmentModel(conn sqlx.SqlConn) SegmentModel {
	return &defaultSegmentModel{
		conn:  conn,
		table: segmentTable,
	}
}

// FindOrCreateByModel 按模型查码段，不存在则按默认值创建（step<=0 归一为 1）。
// 首取并发窗口撞 uk_model 时回查既有码段，保证幂等
func (m *defaultSegmentModel) FindOrCreateByModel(ctx context.Context, modelId, beid int64,
	prefix, datePattern string, step int) (*SegmentRow, error) {
	query := fmt.Sprintf("select %s from %s where `model_id` = ? limit 1", segmentRows, m.table)
	var resp SegmentRow
	err := m.conn.QueryRowCtx(ctx, &resp, query, modelId)
	switch {
	case err == nil:
		return &resp, nil
	case !errors.Is(err, sqlx.ErrNotFound):
		return nil, err
	}

	//不存在则创建（updated_at 由库默认值填充）
	if step <= 0 {
		step = 1
	}
	insertQuery := fmt.Sprintf("insert into %s (`model_id`,`prefix`,`date_pattern`,`current_value`,`step`,`beid`) values (?,?,?,?,?,?)", m.table)
	if _, err := m.conn.ExecCtx(ctx, insertQuery, modelId, prefix, datePattern, 0, step, beid); err != nil {
		if !isDuplicateEntry(err) {
			return nil, err
		}
		//并发首取撞 uk_model：回查既有码段
	}
	var created SegmentRow
	if err := m.conn.QueryRowCtx(ctx, &created, query, modelId); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrSegmentMissing
		}
		return nil, err
	}
	return &created, nil
}

// NextValue 号段取号（并发安全核心）：
// 同一事务内先 update current_value = current_value + step 再 select 回读。
// InnoDB 下 update 对该行加排他锁并持有至事务提交，并发取号在此行锁上串行排队，
// 每个事务读到的都是各自推进后的值，保证不重号、不跳号回退
func (m *defaultSegmentModel) NextValue(ctx context.Context, modelId int64) (*SegmentRow, error) {
	var seg *SegmentRow
	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		//1. 推进流水（获取行锁）
		updateQuery := fmt.Sprintf("update %s set `current_value` = `current_value` + `step` where `model_id` = ?", m.table)
		ret, err := session.ExecCtx(ctx, updateQuery, modelId)
		if err != nil {
			return err
		}
		//推进 0 行：码段不存在，或步长为 0（值未变化）——事务内回查区分
		if affected, _ := ret.RowsAffected(); affected == 0 {
			var probe SegmentRow
			probeQuery := fmt.Sprintf("select %s from %s where `model_id` = ?", segmentRows, m.table)
			switch qErr := session.QueryRowCtx(ctx, &probe, probeQuery, modelId); {
			case qErr == nil:
				return ErrSegmentStepInvalid
			case errors.Is(qErr, sqlx.ErrNotFound):
				return ErrSegmentMissing
			default:
				return qErr
			}
		}
		//2. 事务内回读（读到本事务推进后的值，锁提交前其他事务无法推进）
		var row SegmentRow
		selectQuery := fmt.Sprintf("select %s from %s where `model_id` = ?", segmentRows, m.table)
		if err := session.QueryRowCtx(ctx, &row, selectQuery, modelId); err != nil {
			return err
		}
		seg = &row
		return nil
	})
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrSegmentMissing
		}
		return nil, err
	}
	return seg, nil
}

// DeleteByModel 按模型删除码段（DeleteModel 事务内级联清理用）
func (m *defaultSegmentModel) DeleteByModel(ctx context.Context, session sqlx.Session, modelId, beid int64) error {
	query := fmt.Sprintf("delete from %s where `model_id` = ? and `beid` = ?", m.table)
	_, err := execCtx(ctx, m.conn, session, query, modelId, beid)
	return err
}
