package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	versionTable = "`dc_meta_version`"
)

// versionInsertColumns 插入字段
const versionInsertColumns = "entity_id,version,snapshot_json,operator,published_at"

// versionRows 查询列（snapshot_json 允许 NULL，用 ifnull 兜底避免扫描报错）
const versionRows = "id,entity_id,version,ifnull(snapshot_json,'') as snapshot_json,operator,published_at"

// 元数据版本（定版快照）
type (
	// Version dc_meta_version 行
	Version struct {
		Id           int64     `db:"id"`
		EntityId     int64     `db:"entity_id"`
		Version      int64     `db:"version"`
		SnapshotJson string    `db:"snapshot_json"`
		Operator     string    `db:"operator"`
		PublishedAt  time.Time `db:"published_at"`
	}

	VersionModel interface {
		// Insert 写入版本快照，返回新版本 id
		Insert(ctx context.Context, v *Version) (int64, error)
		// FindOne 按主键查询版本
		FindOne(ctx context.Context, id int64) (*Version, error)
		// FindList 按实体查询版本列表（版本号倒序）
		FindList(ctx context.Context, entityId int64) ([]*Version, error)
		// MaxVersion 查询实体当前最大版本号（无版本时返回 0）
		MaxVersion(ctx context.Context, entityId int64) (int64, error)
	}

	defaultVersionModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewVersionModel 创建元数据版本 model
func NewVersionModel(conn sqlx.SqlConn) VersionModel {
	return &defaultVersionModel{
		conn:  conn,
		table: versionTable,
	}
}

// Insert 写入版本快照
func (m *defaultVersionModel) Insert(ctx context.Context, v *Version) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?,?,?,?,?)", m.table, versionInsertColumns)
	ret, err := m.conn.ExecCtx(ctx, query, v.EntityId, v.Version, v.SnapshotJson, v.Operator, v.PublishedAt)
	if err != nil {
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindOne 按主键查询版本
func (m *defaultVersionModel) FindOne(ctx context.Context, id int64) (*Version, error) {
	query := fmt.Sprintf("select %s from %s where id = ? limit 1", versionRows, m.table)
	var resp Version
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindList 按实体查询版本列表（版本号倒序）
func (m *defaultVersionModel) FindList(ctx context.Context, entityId int64) ([]*Version, error) {
	query := fmt.Sprintf("select %s from %s where entity_id = ? order by version desc", versionRows, m.table)
	var list []*Version
	if err := m.conn.QueryRowsCtx(ctx, &list, query, entityId); err != nil {
		return nil, err
	}
	return list, nil
}

// MaxVersion 查询实体当前最大版本号，无版本时返回 0
func (m *defaultVersionModel) MaxVersion(ctx context.Context, entityId int64) (int64, error) {
	query := fmt.Sprintf("select ifnull(max(version),0) as max_version from %s where entity_id = ?", m.table)
	var maxVersion int64
	if err := m.conn.QueryRowCtx(ctx, &maxVersion, query, entityId); err != nil {
		return 0, err
	}
	return maxVersion, nil
}
