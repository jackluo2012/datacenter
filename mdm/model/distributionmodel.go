package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const distributionTable = "`dc_mdm_distribution`"

// 分发状态（dc_mdm_distribution.status）
const (
	DistStatusPending = "pending"
	DistStatusSuccess = "success"
	DistStatusFailed  = "failed"
)

// distributionRows 查询列（显式 9 列，与 distributionRow 9 字段严格对齐）。
// distributed_at 用 UNIX_TIMESTAMP*1000 输出毫秒时间戳，避免依赖 DSN parseTime 且与 pb int64 对齐
const distributionRows = "`id`,`record_id`,`target_ds_id`,`target_db`,`target_table`,`status`,`msg`,`beid`," +
	"UNIX_TIMESTAMP(`distributed_at`)*1000 AS `distributed_at`"

type (
	// DistributionRow 分发记录查询行：与 distributionRows 的 9 列一一对应
	DistributionRow struct {
		Id            int64  `db:"id"`
		RecordId      int64  `db:"record_id"`
		TargetDsId    int64  `db:"target_ds_id"`
		TargetDb      string `db:"target_db"`
		TargetTable   string `db:"target_table"`
		Status        string `db:"status"` //pending/success/failed
		Msg           string `db:"msg"`
		Beid          int64  `db:"beid"`
		DistributedAt int64  `db:"distributed_at"` //毫秒时间戳
	}

	// DistributionModel 主数据分发记录（dc_mdm_distribution）
	DistributionModel interface {
		// Insert 落一条分发结果（成功/失败均落，msg 截断到 1024 防超列宽）
		Insert(ctx context.Context, data *DistributionRow) (int64, error)
		// FindList 条件分页（beid 必带，recordId 可选），返回列表与总数
		FindList(ctx context.Context, beid, recordId, page, size int64) ([]*DistributionRow, int64, error)
	}

	defaultDistributionModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewDistributionModel 创建分发记录 model
func NewDistributionModel(conn sqlx.SqlConn) DistributionModel {
	return &defaultDistributionModel{
		conn:  conn,
		table: distributionTable,
	}
}

// Insert 落分发结果记录，distributed_at 由库默认值填充为当前时间
func (m *defaultDistributionModel) Insert(ctx context.Context, data *DistributionRow) (int64, error) {
	query := fmt.Sprintf("insert into %s (`record_id`,`target_ds_id`,`target_db`,`target_table`,`status`,`msg`,`beid`) values (?,?,?,?,?,?,?)", m.table)
	ret, err := m.conn.ExecCtx(ctx, query, data.RecordId, data.TargetDsId, data.TargetDb,
		data.TargetTable, data.Status, truncateMsg(data.Msg), data.Beid)
	if err != nil {
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindList 条件分页查询分发记录（recordId<=0 时不限记录），返回列表与总数
func (m *defaultDistributionModel) FindList(ctx context.Context, beid, recordId, page, size int64) ([]*DistributionRow, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	where := "`beid` = ?"
	args := []any{beid}
	if recordId > 0 {
		where += " and `record_id` = ?"
		args = append(args, recordId)
	}

	//总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	//分页
	page, size = normalizePage(page, size)
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?", distributionRows, m.table, where)
	args = append(args, size, (page-1)*size)
	var list []*DistributionRow
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// truncateMsg 截断 msg 至 1024 字节内（列宽 varchar(1024)，按 rune 边界截断避免截出残码）
func truncateMsg(s string) string {
	const limit = 1000 //留余量，避开 utf8mb4 边界误差
	if len(s) <= limit {
		return s
	}
	runes := []rune(s)
	for i := len(runes); i > 0; i-- {
		if len(string(runes[:i])) <= limit {
			return string(runes[:i])
		}
	}
	return ""
}
