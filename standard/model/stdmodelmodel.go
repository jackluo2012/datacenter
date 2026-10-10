package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	stdModelTable           = "`dc_std_model`"
	stdModelInsertFieldSize = 6 // 插入字段数（与 stdModelInsertColumns 对应）
	// stdModelInsertColumns 插入字段（不含自增 id 与时间列；snapshot_json/published_at 仅发布时写）
	stdModelInsertColumns = "`model_code`,`model_name`,`elements_json`,`status`,`version`,`beid`"
	// stdModelRows 查询列（11 列与 StdModel 11 字段一一对应，
	// 规避 go-zero 少列查询扫宽 struct 报 not matching destination to scan；
	// elements_json/snapshot_json 允许 NULL 用 ifnull 兜底，published_at 转毫秒时间戳并以 0 兜底）
	stdModelRows = "`id`,`model_code`,`model_name`,ifnull(`elements_json`,'') as `elements_json`," +
		"`status`,`version`,ifnull(`snapshot_json`,'') as `snapshot_json`," +
		"ifnull(cast(unix_timestamp(`published_at`) * 1000 as signed), 0) as `published_at`," +
		"`beid`,`created_at`,`updated_at`"
)

// 标准模型（dc_std_model，数据元组合成的业务标准）
type (
	// StdModel 标准模型行
	StdModel struct {
		Id           int64     `db:"id"`
		ModelCode    string    `db:"model_code"`
		ModelName    string    `db:"model_name"`
		ElementsJson string    `db:"elements_json"` // 组成元素 [{element_code, required, pk}]
		Status       string    `db:"status"`        // draft/published/deprecated
		Version      int32     `db:"version"`       // 发布版本号，发布 +1
		SnapshotJson string    `db:"snapshot_json"` // 最近发布快照
		PublishedAt  int64     `db:"published_at"`  // 最近发布时间（毫秒时间戳，未发布为 0）
		Beid         int64     `db:"beid"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}

	// StdModelModel 标准模型表操作接口
	StdModelModel interface {
		// Insert 新增标准模型，返回自增 id
		Insert(ctx context.Context, data *StdModel) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*StdModel, error)
		// FindOneByCode 按 beid + 模型编码查询（保存前查重用）
		FindOneByCode(ctx context.Context, beid int64, modelCode string) (*StdModel, error)
		// Update 更新标准模型基础字段（按 id+beid 定位；不动 status/version/快照）
		Update(ctx context.Context, data *StdModel) error
		// Publish 发布：version+1、status=published、落快照与发布时间
		Publish(ctx context.Context, id, beid int64, version int32, snapshotJson string, publishedAt time.Time) error
		// Delete 删除标准模型（按 id+beid 定位）
		Delete(ctx context.Context, id, beid int64) error
		// FindList 条件分页查询：beid 必带，status 可选过滤，keyword 对 model_code/model_name 模糊匹配
		FindList(ctx context.Context, beid int64, status, keyword string, page, size int64) ([]*StdModel, int64, error)
	}

	defaultStdModelModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewStdModelModel 创建标准模型 model
func NewStdModelModel(conn sqlx.SqlConn) StdModelModel {
	return &defaultStdModelModel{
		conn:  conn,
		table: stdModelTable,
	}
}

// Insert 新增标准模型（draft 起步，版本 1）
func (m *defaultStdModelModel) Insert(ctx context.Context, data *StdModel) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, stdModelInsertColumns, placeholders(stdModelInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query,
		data.ModelCode, data.ModelName, data.ElementsJson, data.Status, data.Version, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询标准模型
func (m *defaultStdModelModel) FindOne(ctx context.Context, id, beid int64) (*StdModel, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", stdModelRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp StdModel
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

// FindOneByCode 按 beid + 模型编码查询
func (m *defaultStdModelModel) FindOneByCode(ctx context.Context, beid int64, modelCode string) (*StdModel, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `model_code` = ? limit 1",
		stdModelRows, m.table)
	var resp StdModel
	err := m.conn.QueryRowCtx(ctx, &resp, query, beid, modelCode)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 更新标准模型基础字段（按 id+beid 定位；status/version/快照仅随发布流转）
func (m *defaultStdModelModel) Update(ctx context.Context, data *StdModel) error {
	query := fmt.Sprintf("update %s set `model_code` = ?, `model_name` = ?, `elements_json` = ? "+
		"where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		data.ModelCode, data.ModelName, data.ElementsJson, data.Id, data.Beid)
	return err
}

// Publish 发布标准模型：版本号、快照、发布时间一次落库，status 置 published
func (m *defaultStdModelModel) Publish(ctx context.Context, id, beid int64,
	version int32, snapshotJson string, publishedAt time.Time) error {
	query := fmt.Sprintf("update %s set `status` = ?, `version` = ?, `snapshot_json` = ?, "+
		"`published_at` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		ModelStatusPublished, version, snapshotJson, publishedAt, id, beid)
	return err
}

// Delete 删除标准模型
func (m *defaultStdModelModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 标准模型列表（beid 多租户必带 + status 可选过滤 + keyword 对 model_code/model_name 模糊匹配）
func (m *defaultStdModelModel) FindList(ctx context.Context, beid int64, status, keyword string,
	page, size int64) ([]*StdModel, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if status != "" {
		where += " and `status` = ?"
		args = append(args, status)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and (`model_code` like ? or `model_name` like ?)"
		like := "%" + escapeLike(kw) + "%"
		args = append(args, like, like)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		stdModelRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*StdModel
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
