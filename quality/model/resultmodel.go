package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	resultTable           = "`dc_quality_result`"
	resultInsertFieldSize = 7   // 插入字段数（与 resultInsertColumns 对应）
	resultInsertChunk     = 500 // 批量插入分片大小
	// resultInsertColumns 插入字段（不含自增 id，run_at 显式传入保证批量运行时间一致）
	resultInsertColumns = "job_id,rule_id,rule_type,check_value,pass_flag,detail_json,run_at"
	// resultRows 查询列（detail_json 允许 NULL 用 ifnull 兜底；run_at 转毫秒时间戳并以 0 兜底，
	// 查询列数与 Result 字段数一一对应，规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	resultRows = "id,job_id,rule_id,rule_type,check_value,pass_flag,ifnull(detail_json,'') as detail_json," +
		"ifnull(cast(unix_timestamp(run_at) * 1000 as signed), 0) as run_at"
)

// 质量检查结果（dc_quality_result）
type (
	// Result 质量检查结果行
	Result struct {
		Id         int64  `db:"id"`
		JobId      int64  `db:"job_id"`
		RuleId     int64  `db:"rule_id"`
		RuleType   string `db:"rule_type"`
		CheckValue string `db:"check_value"` // 检查实测值
		PassFlag   int64  `db:"pass_flag"`   // 1.通过 0.不通过
		DetailJson string `db:"detail_json"` // 详情(阈值/实测/差异/错误原因)
		RunAt      int64  `db:"run_at"`      // 执行时间（毫秒时间戳）
	}

	// ResultModel 质量结果表操作接口
	ResultModel interface {
		// BatchInsert 批量写入检查结果（一次任务的所有规则结果一次落库）
		BatchInsert(ctx context.Context, results []*Result) error
		// FindListByJob 按任务分页查询结果：passFilter 0.全部 1.仅通过 2.仅不通过；返回列表与总数
		FindListByJob(ctx context.Context, jobId, passFilter, page, size int64) ([]*Result, int64, error)
	}

	defaultResultModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewResultModel 创建质量结果 model
func NewResultModel(conn sqlx.SqlConn) ResultModel {
	return &defaultResultModel{
		conn:  conn,
		table: resultTable,
	}
}

// BatchInsert 批量写入检查结果：分片拼接多行 values
func (m *defaultResultModel) BatchInsert(ctx context.Context, results []*Result) error {
	if len(results) == 0 {
		return nil
	}
	prefix := fmt.Sprintf("insert into %s (%s) values ", m.table, resultInsertColumns)
	for start := 0; start < len(results); start += resultInsertChunk {
		end := start + resultInsertChunk
		if end > len(results) {
			end = len(results)
		}
		batch := results[start:end]
		query := prefix + strings.TrimSuffix(strings.Repeat("("+placeholders(resultInsertFieldSize)+"),", len(batch)), ",")
		args := make([]any, 0, len(batch)*resultInsertFieldSize)
		for _, r := range batch {
			args = append(args, r.JobId, r.RuleId, r.RuleType, r.CheckValue, r.PassFlag, r.DetailJson, time.UnixMilli(r.RunAt))
		}
		if _, err := m.conn.ExecCtx(ctx, query, args...); err != nil {
			return fmt.Errorf("批量写入质量结果失败: %w", err)
		}
	}
	return nil
}

// FindListByJob 按任务分页查询结果（passFilter 0.全部 1.仅通过 2.仅不通过），返回列表与总数
func (m *defaultResultModel) FindListByJob(ctx context.Context, jobId, passFilter, page, size int64) ([]*Result, int64, error) {
	if jobId <= 0 {
		return nil, 0, ErrJobIdRequired
	}
	page, size = normalizePage(page, size)

	where := "`job_id` = ?"
	args := []any{jobId}
	switch passFilter {
	case 1: // 仅通过
		where += " and `pass_flag` = 1"
	case 2: // 仅不通过
		where += " and `pass_flag` = 0"
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		resultRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Result
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
