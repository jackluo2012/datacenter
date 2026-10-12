package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	// 注意坑：builder.RawFieldNames 返回带反引号列名，stringx.Remove 必须用 "`id`" 形式
	tagDefFieldNames          = builder.RawFieldNames(&TagDef{})
	tagDefRowsExpectAutoSet   = strings.Join(stringx.Remove(tagDefFieldNames, "`id`", "`created_at`", "`updated_at`"), ",")
	tagDefRowsWithPlaceHolder = strings.Join(stringx.Remove(tagDefFieldNames, "`id`", "`created_at`", "`updated_at`"), "=?,") + "=?"
)

const (
	tagDefTable           = "`dc_tag_def`"
	tagDefInsertFieldSize = 8 // 插入字段数（与 tagDefRowsExpectAutoSet 对应）
	tagDefDefaultPageSize = 10
	tagDefMaxPageSize     = 200
)

// tagDefRows 查询列（rule_sql 为 text 可空，用 ifnull 兜底避免扫描报错）；
// 列顺序与 TagDef 字段顺序一致（11 列对 11 字段，列数不一致扫宽 struct 会报错）
const tagDefRows = "`id`,`object_id`,`tag_key`,`tag_name`,`tag_type`,ifnull(`rule_sql`,'') as `rule_sql`,`refresh_cron`,`status`,`beid`,`created_at`,`updated_at`"

type (
	// TagDef dc_tag_def 行（标签定义）
	TagDef struct {
		Id          int64     `db:"id"`
		ObjectId    int64     `db:"object_id"`
		TagKey      string    `db:"tag_key"`
		TagName     string    `db:"tag_name"`
		TagType     string    `db:"tag_type"` //rule规则/fact事实/model模型
		RuleSql     string    `db:"rule_sql"` //select <pk> as object_id, <expr> as tag_value from ...
		RefreshCron string    `db:"refresh_cron"`
		Status      int32     `db:"status"` //1.启用 0.停用
		Beid        int64     `db:"beid"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

	TagDefModel interface {
		// Create 新增标签定义（object_id+tag_key+beid 唯一查重），返回自增 id
		Create(ctx context.Context, data *TagDef) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*TagDef, error)
		// FindList 分页列表（beid 必带，objectId 可选），返回列表与总数
		FindList(ctx context.Context, beid, objectId, page, size int64) ([]*TagDef, int64, error)
		// FindEnabledRuleByObject 对象下全部启用中（status=1）的 rule 标签（批量打标用）
		FindEnabledRuleByObject(ctx context.Context, objectId, beid int64) ([]*TagDef, error)
		// Update 全量更新（不含自增 id 与时间列）
		Update(ctx context.Context, data *TagDef) error
		// Delete 删除定义并级联删除其全部标签值（同一事务内先删值后删定义）
		Delete(ctx context.Context, id, beid int64) error
		// FindOneByKey 按对象+标签键查询（beid>0 时追加多租户过滤），更新换 key 查重用
		FindOneByKey(ctx context.Context, objectId int64, tagKey string, beid int64) (*TagDef, error)
		// CountByObject 统计对象下的标签定义数（删除对象前引用校验用）
		CountByObject(ctx context.Context, objectId, beid int64) (int64, error)
	}

	defaultTagDefModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewTagDefModel 创建标签定义 model
func NewTagDefModel(conn sqlx.SqlConn) TagDefModel {
	return &defaultTagDefModel{
		conn:  conn,
		table: tagDefTable,
	}
}

// Create 新增标签定义：先按 object_id+tag_key+beid 查重，命中即拒绝；
// 并发窗口撞 uk_obj_key 唯一键时按重复键错误统一处理
func (m *defaultTagDefModel) Create(ctx context.Context, data *TagDef) (int64, error) {
	dupQuery := fmt.Sprintf("select `id` from %s where `object_id` = ? and `tag_key` = ? and `beid` = ? limit 1", m.table)
	var existId int64
	err := m.conn.QueryRowCtx(ctx, &existId, dupQuery, data.ObjectId, data.TagKey, data.Beid)
	switch {
	case err == nil:
		return 0, ErrDupTagKey
	case !errors.Is(err, sqlx.ErrNotFound):
		return 0, err
	}

	query := fmt.Sprintf("insert into %s (%s) values (%s)", m.table, tagDefRowsExpectAutoSet, placeholders(tagDefInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query, data.ObjectId, data.TagKey, data.TagName, data.TagType,
		data.RuleSql, data.RefreshCron, data.Status, data.Beid)
	if err != nil {
		if IsDuplicateEntry(err) {
			return 0, ErrDupTagKey
		}
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindOne 按主键查询标签定义
func (m *defaultTagDefModel) FindOne(ctx context.Context, id, beid int64) (*TagDef, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", tagDefRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp TagDef
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

// FindList 分页查询标签定义，objectId<=0 时查全部对象
func (m *defaultTagDefModel) FindList(ctx context.Context, beid, objectId, page, size int64) ([]*TagDef, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	where := "`beid` = ?"
	args := []any{beid}
	if objectId > 0 {
		where += " and `object_id` = ?"
		args = append(args, objectId)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页参数防御
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = tagDefDefaultPageSize
	}
	if size > tagDefMaxPageSize {
		size = tagDefMaxPageSize
	}
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?", tagDefRows, m.table, where)
	args = append(args, size, (page-1)*size)
	var list []*TagDef
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindEnabledRuleByObject 查对象下全部启用中的 rule 标签（tag_id=0 批量打标用），按 id 升序
func (m *defaultTagDefModel) FindEnabledRuleByObject(ctx context.Context, objectId, beid int64) ([]*TagDef, error) {
	query := fmt.Sprintf("select %s from %s where `object_id` = ? and `tag_type` = 'rule' and `status` = 1", tagDefRows, m.table)
	args := []any{objectId}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var list []*TagDef
	if err := m.conn.QueryRowsCtx(ctx, &list, query+" order by `id` asc", args...); err != nil {
		return nil, err
	}
	return list, nil
}

// Update 全量更新标签定义（带多租户过滤）
func (m *defaultTagDefModel) Update(ctx context.Context, data *TagDef) error {
	query := fmt.Sprintf("update %s set %s where `id` = ? and `beid` = ?", m.table, tagDefRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.ObjectId, data.TagKey, data.TagName, data.TagType,
		data.RuleSql, data.RefreshCron, data.Status, data.Beid, data.Id, data.Beid)
	return err
}

// Delete 删除标签定义并级联删除其全部标签值：同一事务内先删 dc_tag_value 再删 dc_tag_def，
// 保证不留孤儿标签值
func (m *defaultTagDefModel) Delete(ctx context.Context, id, beid int64) error {
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先级联删除该标签全部打标值
		delValues := "delete from `dc_tag_value` where `tag_id` = ?"
		if _, err := session.ExecCtx(ctx, delValues, id); err != nil {
			return fmt.Errorf("级联删除标签值失败: %w", err)
		}
		// 再删标签定义
		delDef := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
		if _, err := session.ExecCtx(ctx, delDef, id, beid); err != nil {
			return fmt.Errorf("删除标签定义失败: %w", err)
		}
		return nil
	})
}

// FindOneByKey 按对象+标签键查询标签定义（唯一键 uk_obj_key 口径），未找到返回 ErrNotFound
func (m *defaultTagDefModel) FindOneByKey(ctx context.Context, objectId int64, tagKey string, beid int64) (*TagDef, error) {
	query := fmt.Sprintf("select %s from %s where `object_id` = ? and `tag_key` = ?", tagDefRows, m.table)
	args := []any{objectId, tagKey}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp TagDef
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

// CountByObject 统计对象下的标签定义数，>0 表示对象仍被标签引用不可删
func (m *defaultTagDefModel) CountByObject(ctx context.Context, objectId, beid int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `object_id` = ?", m.table)
	args := []any{objectId}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var n int64
	err := m.conn.QueryRowCtx(ctx, &n, query, args...)
	return n, err
}
