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
	categoryFieldNames          = builder.RawFieldNames(&AssetCategory{})
	categoryRows                = strings.Join(categoryFieldNames, ",")
	categoryRowsExpectAutoSet   = strings.Join(stringx.Remove(categoryFieldNames, "`id`", "`created_at`"), ",")
	categoryRowsWithPlaceHolder = strings.Join(stringx.Remove(categoryFieldNames, "`id`", "`created_at`"), "=?,") + "=?"
)

type (
	// AssetCategory 资产目录分类表 dc_asset_category
	AssetCategory struct {
		Id        int64     `db:"id"` // 分类id
		ParentId  int64     `db:"parent_id"`
		Name      string    `db:"name"`
		Beid      int64     `db:"beid"` // 平台id
		CreatedAt time.Time `db:"created_at"`
	}

	// CategoryModel 分类表操作接口
	CategoryModel interface {
		Insert(data AssetCategory) (sql.Result, error)
		FindOne(id, beid int64) (*AssetCategory, error)
		Update(data AssetCategory) error
		// Delete 软删分类（置 deleted_at，带平台隔离；重复删除幂等）
		Delete(id, beid int64) error
		// CountChildren 统计分类下未删除的子分类数（删除/挂父前引用校验用）
		CountChildren(id, beid int64) (int64, error)
		// ListByBeid 按平台查询全部未删除分类，parent_id 升序排列，便于前端组树
		ListByBeid(beid int64) ([]*AssetCategory, error)
	}

	defaultCategoryModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewCategoryModel 创建分类 model
func NewCategoryModel(conn sqlx.SqlConn) CategoryModel {
	return &defaultCategoryModel{
		conn:  conn,
		table: "`dc_asset_category`",
	}
}

// Insert 新增分类
func (m *defaultCategoryModel) Insert(data AssetCategory) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?)", m.table, categoryRowsExpectAutoSet)
	ret, err := m.conn.Exec(query, data.ParentId, data.Name, data.Beid)
	return ret, err
}

// FindOne 按分类id查询未删除记录（带平台隔离）
func (m *defaultCategoryModel) FindOne(id, beid int64) (*AssetCategory, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? and `beid` = ? and `deleted_at` is null limit 1", categoryRows, m.table)
	var resp AssetCategory
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

// Update 更新分类（含 parent_id/name，带平台隔离）
func (m *defaultCategoryModel) Update(data AssetCategory) error {
	query := fmt.Sprintf("update %s set %s where `id` = ? and `beid` = ?", m.table, categoryRowsWithPlaceHolder)
	_, err := m.conn.Exec(query, data.ParentId, data.Name, data.Beid, data.Id, data.Beid)
	return err
}

// Delete 软删分类：置 deleted_at 而非物理删除，保留历史与审计线索；
// 带 deleted_at is null 条件保证重复删除幂等
func (m *defaultCategoryModel) Delete(id, beid int64) error {
	query := fmt.Sprintf("update %s set `deleted_at` = now(3) where `id` = ? and `beid` = ? and `deleted_at` is null", m.table)
	_, err := m.conn.Exec(query, id, beid)
	return err
}

// CountChildren 统计分类下未删除的子分类数，>0 表示有子分类不可删
func (m *defaultCategoryModel) CountChildren(id, beid int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `parent_id` = ? and `beid` = ? and `deleted_at` is null", m.table)
	var n int64
	err := m.conn.QueryRow(&n, query, id, beid)
	return n, err
}

// ListByBeid 查询平台下全部未删除分类，parent_id 升序（同级再按 id 升序）
func (m *defaultCategoryModel) ListByBeid(beid int64) ([]*AssetCategory, error) {
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `deleted_at` is null order by `parent_id` asc, `id` asc", categoryRows, m.table)
	var resp []*AssetCategory
	err := m.conn.QueryRowsPartial(&resp, query, beid)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
