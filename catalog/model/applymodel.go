package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	applyFieldNames          = builder.RawFieldNames(&AssetApply{})
	applyRows                = strings.Join(applyFieldNames, ",")
	applyRowsExpectAutoSet   = strings.Join(stringx.Remove(applyFieldNames, "`id`", "`created_at`", "`updated_at`"), ",")
	applyRowsWithPlaceHolder = strings.Join(stringx.Remove(applyFieldNames, "`id`", "`created_at`", "`updated_at`"), "=?,") + "=?"
)

// 分页默认值与上限
const (
	applyDefaultPageSize = 10
	applyMaxPageSize     = 200
)

type (
	// AssetApply 资产申请使用表 dc_asset_apply
	AssetApply struct {
		Id        int64     `db:"id"`        // 申请id
		AssetId   int64     `db:"asset_id"`  // 资产id
		Reason    string    `db:"reason"`    // 申请理由
		Applicant string    `db:"applicant"` // 申请人
		Status    string    `db:"status"`    // pending/approved/rejected
		Reviewer  string    `db:"reviewer"`  // 审批人
		Beid      int64     `db:"beid"`      // 平台id
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	// ApplyModel 资产申请表操作接口
	ApplyModel interface {
		Insert(data AssetApply) (sql.Result, error)
		FindOne(id, beid int64) (*AssetApply, error)
		Update(data AssetApply) error
		// UpdateStatus 审批落库：仅当当前状态仍为 pending 时生效（乐观并发控制），
		// 返回受影响行数，0 表示已被他人审批
		UpdateStatus(id, beid int64, status, reviewer string) (int64, error)
		// ListByBeid 申请列表：beid 必带，status 非空则过滤；返回列表+总数
		ListByBeid(beid int64, status string, page, size int64) ([]*AssetApply, int64, error)
	}

	defaultApplyModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewApplyModel 创建资产申请 model
func NewApplyModel(conn sqlx.SqlConn) ApplyModel {
	return &defaultApplyModel{
		conn:  conn,
		table: "`dc_asset_apply`",
	}
}

// Insert 新增申请
func (m *defaultApplyModel) Insert(data AssetApply) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?)", m.table, applyRowsExpectAutoSet)
	ret, err := m.conn.Exec(query, data.AssetId, data.Reason, data.Applicant, data.Status, data.Reviewer, data.Beid)
	return ret, err
}

// FindOne 按申请id查询（带平台隔离）
func (m *defaultApplyModel) FindOne(id, beid int64) (*AssetApply, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? and `beid` = ? limit 1", applyRows, m.table)
	var resp AssetApply
	err := m.conn.QueryRowPartial(&resp, query, id, beid)
	switch {
	case err == nil:
		return &resp, nil
	case err == sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 全量更新申请（带平台隔离）
func (m *defaultApplyModel) Update(data AssetApply) error {
	query := fmt.Sprintf("update %s set %s where `id` = ? and `beid` = ?", m.table, applyRowsWithPlaceHolder)
	_, err := m.conn.Exec(query, data.AssetId, data.Reason, data.Applicant, data.Status, data.Reviewer, data.Beid,
		data.Id, data.Beid)
	return err
}

// UpdateStatus 审批落库：带 `status = 'pending'` 条件防止并发重复审批
func (m *defaultApplyModel) UpdateStatus(id, beid int64, status, reviewer string) (int64, error) {
	query := fmt.Sprintf("update %s set `status` = ?, `reviewer` = ? where `id` = ? and `beid` = ? and `status` = 'pending'", m.table)
	ret, err := m.conn.Exec(query, status, reviewer, id, beid)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

// ListByBeid 申请列表（分页 + 可选状态过滤），返回列表与总数
func (m *defaultApplyModel) ListByBeid(beid int64, status string, page, size int64) ([]*AssetApply, int64, error) {
	// 分页参数防御
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = applyDefaultPageSize
	}
	if size > applyMaxPageSize {
		size = applyMaxPageSize
	}

	// 动态拼接查询条件（beid 多租户必带）
	conds := []string{"`beid` = ?"}
	args := []interface{}{beid}
	if status != "" {
		conds = append(conds, "`status` = ?")
		args = append(args, status)
	}
	where := strings.Join(conds, " and ")

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowPartial(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?", applyRows, m.table, where)
	listArgs := append(append([]interface{}{}, args...), size, (page-1)*size)
	var resp []*AssetApply
	if err := m.conn.QueryRowsPartial(&resp, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}
