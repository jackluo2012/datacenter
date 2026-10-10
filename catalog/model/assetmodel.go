package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	assetFieldNames          = builder.RawFieldNames(&Asset{})
	assetRows                = strings.Join(assetFieldNames, ",")
	assetRowsExpectAutoSet   = strings.Join(stringx.Remove(assetFieldNames, "`id`", "`created_at`", "`updated_at`"), ",")
	assetRowsWithPlaceHolder = strings.Join(stringx.Remove(assetFieldNames, "`id`", "`created_at`", "`updated_at`"), "=?,") + "=?"
)

// 分页默认值与上限
const (
	assetDefaultPageSize = 10
	assetMaxPageSize     = 200
)

type (
	// Asset 资产表 dc_asset（元数据实体入目录，冗余快照避免跨服务联查）
	Asset struct {
		Id             int64     `db:"id"`        // 资产id
		EntityId       int64     `db:"entity_id"` // 元数据实体id
		CategoryId     int64     `db:"category_id"`
		Name           string    `db:"name"`            // 资产名(冗余快照)
		EntityType     string    `db:"entity_type"`     // 冗余快照
		Comment        string    `db:"comment"`         // 冗余快照
		Classification string    `db:"classification"`  // 冗余快照
		AssetTagsJson  string    `db:"asset_tags_json"` // 资产标签(json)
		Visibility     string    `db:"visibility"`      // public/private
		Beid           int64     `db:"beid"`            // 平台id
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	// AssetModel 资产表操作接口
	AssetModel interface {
		// TagAsset 登记资产（upsert）：uk_beid_entity 冲突时刷新分类/标签/可见性及冗余快照
		TagAsset(data Asset) (int64, error)
		FindOne(id, beid int64) (*Asset, error)
		Update(data Asset) error
		Delete(id, beid int64) error
		// FindList 资产列表：beid 必带，categoryId>0 过滤分类，keyword 模糊匹配 name/comment；返回列表+总数
		FindList(beid, categoryId int64, keyword string, page, size int64) ([]*Asset, int64, error)
	}

	defaultAssetModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewAssetModel 创建资产 model
func NewAssetModel(conn sqlx.SqlConn) AssetModel {
	return &defaultAssetModel{
		conn:  conn,
		table: "`dc_asset`",
	}
}

// TagAsset 登记资产（upsert）。
// 唯一键 uk_beid_entity 冲突时更新分类/标签/可见性及其余冗余快照字段，
// 并通过 last_insert_id(id) 保证冲突更新时也能拿到既有资产id。
func (m *defaultAssetModel) TagAsset(data Asset) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?) "+
		"on duplicate key update `id` = last_insert_id(`id`), `category_id` = values(`category_id`), "+
		"`name` = values(`name`), `entity_type` = values(`entity_type`), `comment` = values(`comment`), "+
		"`classification` = values(`classification`), `asset_tags_json` = values(`asset_tags_json`), "+
		"`visibility` = values(`visibility`)", m.table, assetRowsExpectAutoSet)
	ret, err := m.conn.Exec(query, data.EntityId, data.CategoryId, data.Name, data.EntityType,
		data.Comment, data.Classification, data.AssetTagsJson, data.Visibility, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按资产id查询（带平台隔离）
func (m *defaultAssetModel) FindOne(id, beid int64) (*Asset, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? and `beid` = ? limit 1", assetRows, m.table)
	var resp Asset
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

// Update 更新资产（带平台隔离）
func (m *defaultAssetModel) Update(data Asset) error {
	query := fmt.Sprintf("update %s set %s where `id` = ? and `beid` = ?", m.table, assetRowsWithPlaceHolder)
	_, err := m.conn.Exec(query, data.EntityId, data.CategoryId, data.Name, data.EntityType,
		data.Comment, data.Classification, data.AssetTagsJson, data.Visibility, data.Beid,
		data.Id, data.Beid)
	return err
}

// Delete 删除资产（带平台隔离）
func (m *defaultAssetModel) Delete(id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.Exec(query, id, beid)
	return err
}

// FindList 资产列表（分页 + 分类过滤 + keyword 模糊匹配 name/comment），返回列表与总数
func (m *defaultAssetModel) FindList(beid, categoryId int64, keyword string, page, size int64) ([]*Asset, int64, error) {
	// 分页参数防御
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = assetDefaultPageSize
	}
	if size > assetMaxPageSize {
		size = assetMaxPageSize
	}

	// 动态拼接查询条件（beid 多租户必带，参数化防注入）
	conds := []string{"`beid` = ?"}
	args := []interface{}{beid}
	if categoryId > 0 {
		conds = append(conds, "`category_id` = ?")
		args = append(args, categoryId)
	}
	if keyword != "" {
		conds = append(conds, "(`name` like ? or `comment` like ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}
	where := strings.Join(conds, " and ")

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowPartial(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?", assetRows, m.table, where)
	listArgs := append(append([]interface{}{}, args...), size, (page-1)*size)
	var resp []*Asset
	if err := m.conn.QueryRowsPartial(&resp, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}
