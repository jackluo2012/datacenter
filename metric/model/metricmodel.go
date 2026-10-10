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

const (
	metricTable      = "`dc_metric`"
	metricInsertSize = 11  // 插入字段数（与 metricInsertColumns 对应）
	metricIdChunk    = 500 // id in 查询分片大小
	metricPageSize   = 10  // 默认分页大小
)

// 指标类型
const (
	MetricTypeAtom   = "atom"   // 原子指标
	MetricTypeDerive = "derive" // 派生指标
)

var (
	// metricFieldNames RawFieldNames 返回带反引号的列名（如 `id`），stringx.Remove 移除时必须同样带反引号
	metricFieldNames = builder.RawFieldNames(&Metric{})
	// metricInsertColumns 插入列（不含自增 id 与时间列）
	metricInsertColumns = strings.Join(stringx.Remove(metricFieldNames, "`id`", "`created_at`", "`updated_at`"), ",")
)

// metricRows 查询列（与 Metric 字段一一对应，列数必须与 struct 字段数一致，
// 否则少列查询扫宽 struct 会报 not matching destination to scan；
// dimension_json 允许 NULL，用 ifnull 兜底避免扫描报错）
const metricRows = "`id`,`name_en`,`name_cn`,`metric_type`,`biz_field`,`agg_type`,`expr`," +
	"ifnull(dimension_json,'') as dimension_json,`comment`,`owner`,`status`,`beid`,`created_at`,`updated_at`"

// 数据指标（atom 原子指标 / derive 派生指标）
type (
	// Metric dc_metric 行
	Metric struct {
		Id            int64     `db:"id"`
		NameEn        string    `db:"name_en"`
		NameCn        string    `db:"name_cn"`
		MetricType    string    `db:"metric_type"`
		BizField      string    `db:"biz_field"`
		AggType       string    `db:"agg_type"`
		Expr          string    `db:"expr"`
		DimensionJson string    `db:"dimension_json"`
		Comment       string    `db:"comment"`
		Owner         string    `db:"owner"`
		Status        int64     `db:"status"`
		Beid          int64     `db:"beid"`
		CreatedAt     time.Time `db:"created_at"`
		UpdatedAt     time.Time `db:"updated_at"`
	}

	MetricModel interface {
		// Insert 写入指标（uk_beid_name: name_en+beid 唯一），返回新指标 id
		Insert(ctx context.Context, data *Metric) (int64, error)
		// FindOneByNameEn 按 beid + 英文名查询（创建前查重）
		FindOneByNameEn(ctx context.Context, beid int64, nameEn string) (*Metric, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Metric, error)
		// FindList 条件分页查询（metric_type/keyword 均可选，keyword 对 name_en/name_cn 模糊匹配），返回列表与总数
		FindList(ctx context.Context, beid int64, metricType, keyword string, page, size int64) ([]*Metric, int64, error)
		// FindListByIds 按主键批量查询（自动去重并分片，beid>0 时追加多租户过滤）
		FindListByIds(ctx context.Context, beid int64, ids []int64) ([]*Metric, error)
		// Update 部分更新：仅覆盖非空字段（name_cn/expr/dimension_json/comment）
		Update(ctx context.Context, data *Metric) error
		// UpdateStatus 更新指标状态
		UpdateStatus(ctx context.Context, id, beid int64, status int64) error
		// Delete 按主键删除（session 可空，级联删 rel 时随事务传入）
		Delete(ctx context.Context, session sqlx.Session, id, beid int64) error
	}

	defaultMetricModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewMetricModel 创建数据指标 model
func NewMetricModel(conn sqlx.SqlConn) MetricModel {
	return &defaultMetricModel{
		conn:  conn,
		table: metricTable,
	}
}

// Insert 写入指标，返回自增主键
func (m *defaultMetricModel) Insert(ctx context.Context, data *Metric) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)", m.table, metricInsertColumns, placeholders(metricInsertSize))
	ret, err := m.conn.ExecCtx(ctx, query, data.NameEn, data.NameCn, data.MetricType, data.BizField,
		data.AggType, data.Expr, data.DimensionJson, data.Comment, data.Owner, data.Status, data.Beid)
	if err != nil {
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindOneByNameEn 按 beid + 英文名查询指标（uk_beid_name 查重）
func (m *defaultMetricModel) FindOneByNameEn(ctx context.Context, beid int64, nameEn string) (*Metric, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `name_en` = ? limit 1", metricRows, m.table)
	var resp Metric
	err := m.conn.QueryRowCtx(ctx, &resp, query, beid, nameEn)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindOne 按主键查询指标
func (m *defaultMetricModel) FindOne(ctx context.Context, id, beid int64) (*Metric, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", metricRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Metric
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

// FindList 条件分页查询指标，返回列表与总数
func (m *defaultMetricModel) FindList(ctx context.Context, beid int64, metricType, keyword string,
	page, size int64) ([]*Metric, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	where := "`beid` = ?"
	args := []any{beid}
	if metricType != "" {
		where += " and `metric_type` = ?"
		args = append(args, metricType)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and (`name_en` like ? or `name_cn` like ?)"
		like := "%" + escapeLike(kw) + "%"
		args = append(args, like, like)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) as total from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = metricPageSize
	}
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?", metricRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Metric
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindListByIds 按主键批量查询指标（自动去重并分片）
func (m *defaultMetricModel) FindListByIds(ctx context.Context, beid int64, ids []int64) ([]*Metric, error) {
	uniq := UniqInt64s(ids)
	if len(uniq) == 0 {
		return []*Metric{}, nil
	}

	list := make([]*Metric, 0, len(uniq))
	for start := 0; start < len(uniq); start += metricIdChunk {
		end := start + metricIdChunk
		if end > len(uniq) {
			end = len(uniq)
		}
		part := uniq[start:end]
		query := fmt.Sprintf("select %s from %s where `id` in (%s)", metricRows, m.table, placeholders(len(part)))
		args := make([]any, 0, len(part)+1)
		for _, id := range part {
			args = append(args, id)
		}
		if beid > 0 {
			query += " and `beid` = ?"
			args = append(args, beid)
		}
		var rows []*Metric
		if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
			return nil, err
		}
		list = append(list, rows...)
	}
	return list, nil
}

// Update 部分更新指标：动态拼接非空字段（name_cn/expr/dimension_json/comment），空字段保持原值
func (m *defaultMetricModel) Update(ctx context.Context, data *Metric) error {
	sets := make([]string, 0, 4)
	args := make([]any, 0, 6)
	if data.NameCn != "" {
		sets = append(sets, "`name_cn` = ?")
		args = append(args, data.NameCn)
	}
	if data.Expr != "" {
		sets = append(sets, "`expr` = ?")
		args = append(args, data.Expr)
	}
	if data.DimensionJson != "" {
		sets = append(sets, "`dimension_json` = ?")
		args = append(args, data.DimensionJson)
	}
	if data.Comment != "" {
		sets = append(sets, "`comment` = ?")
		args = append(args, data.Comment)
	}
	if len(sets) == 0 {
		// 无可更新字段，直接返回，避免生成非法 SQL
		return nil
	}
	query := fmt.Sprintf("update %s set %s where `id` = ? and `beid` = ?", m.table, strings.Join(sets, ","))
	args = append(args, data.Id, data.Beid)
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// UpdateStatus 更新指标状态（1 启用 / 0 停用）
func (m *defaultMetricModel) UpdateStatus(ctx context.Context, id, beid int64, status int64) error {
	query := fmt.Sprintf("update %s set `status` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id, beid)
	return err
}

// Delete 按主键删除指标（session 可空：派生指标级联删 rel 时随事务传入）
func (m *defaultMetricModel) Delete(ctx context.Context, session sqlx.Session, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := execCtx(ctx, m.conn, session, query, id, beid)
	return err
}
