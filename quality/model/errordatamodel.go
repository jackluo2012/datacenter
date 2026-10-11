package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	errorDataTable           = "`dc_quality_error_data`"
	errorDataInsertFieldSize = 5 // 插入字段数（与 errorDataInsertColumns 对应）
	// errorDataInsertColumns 插入字段（不含自增 id，run_at 显式传入保证与任务执行时间对齐）
	errorDataInsertColumns = "rule_id,job_id,rows_json,sample_count,run_at"
	// errorDataRows 查询列（rows_json 允许 NULL 用 ifnull 兜底；run_at 转毫秒时间戳并以 0 兜底，
	// 查询列数与 ErrorData 字段数一一对应，规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	errorDataRows = "id,rule_id,job_id,ifnull(rows_json,'') as rows_json,sample_count," +
		"ifnull(cast(unix_timestamp(run_at) * 1000 as signed), 0) as run_at"
)

// 错误数据样本（dc_quality_error_data）
type (
	// ErrorData 错误数据样本行（一次任务一条失败规则最多存 100 行样本）
	ErrorData struct {
		Id          int64  `db:"id"`
		RuleId      int64  `db:"rule_id"`
		JobId       int64  `db:"job_id"`
		RowsJson    string `db:"rows_json"`    // 违规行样本 json 数组
		SampleCount int64  `db:"sample_count"` // 实抓样本行数
		RunAt       int64  `db:"run_at"`       // 抓取时间（毫秒时间戳）
	}

	// ErrorDataModel 错误数据样本表操作接口
	ErrorDataModel interface {
		// Insert 写入一次错误样本（每条失败规则一条记录）
		Insert(ctx context.Context, data *ErrorData) error
		// FindList 错误样本分页：jobId 必带，ruleId>0 时追加过滤；返回列表与总数
		FindList(ctx context.Context, jobId, ruleId, page, size int64) ([]*ErrorData, int64, error)
	}

	defaultErrorDataModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewErrorDataModel 创建错误数据样本 model
func NewErrorDataModel(conn sqlx.SqlConn) ErrorDataModel {
	return &defaultErrorDataModel{
		conn:  conn,
		table: errorDataTable,
	}
}

// Insert 写入一次错误样本
func (m *defaultErrorDataModel) Insert(ctx context.Context, data *ErrorData) error {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, errorDataInsertColumns, placeholders(errorDataInsertFieldSize))
	_, err := m.conn.ExecCtx(ctx, query,
		data.RuleId, data.JobId, data.RowsJson, data.SampleCount, time.UnixMilli(data.RunAt))
	return err
}

// FindList 错误样本分页（job_id 必带 + rule_id 可选过滤），返回列表与总数
func (m *defaultErrorDataModel) FindList(ctx context.Context, jobId, ruleId, page, size int64) ([]*ErrorData, int64, error) {
	if jobId <= 0 {
		return nil, 0, ErrJobIdRequired
	}
	page, size = normalizePage(page, size)

	where := "`job_id` = ?"
	args := []any{jobId}
	if ruleId > 0 {
		where += " and `rule_id` = ?"
		args = append(args, ruleId)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		errorDataRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*ErrorData
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
