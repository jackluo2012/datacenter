package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"datacenter/quality/model"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ErrNoErrorSample 该规则类型无"违规行"语义（row_count/fluctuation 报行数、consistency 报引用缺失，
// 样本下钻仅覆盖 not_null/unique/value_range）
var ErrNoErrorSample = errors.New("该规则类型不支持抓取错误数据样本")

// errorSampleLimit 单条失败规则最多抓取的违规行样本数
const errorSampleLimit = 100

// errorDataToPb 错误样本 model -> pb（sample_count 为 int32 收窄转换）
func errorDataToPb(e *model.ErrorData) *quality.ErrorData {
	return &quality.ErrorData{
		RuleId:      e.RuleId,
		JobId:       e.JobId,
		RowsJson:    e.RowsJson,
		SampleCount: int32(e.SampleCount),
		RunAt:       e.RunAt,
	}
}

// BuildErrorSampleSQL 生成抓取违规行样本的 SQL（纯函数，固定 limit 100）。
// 表名/列名先过标识符白名单，数值参数走占位符（返回 args）。各类型 SQL：
//
//	not_null    → select * from <table> where <col> is null or <col>='' limit 100
//	unique      → select t.* from <table> t join
//	              (select <col> from <table> group by <col> having count(*)>1) dup
//	              on t.<col> = dup.<col> limit 100（重复组的全部行）
//	value_range → select * from <table> where <col> not between ? and ? limit 100（越界行）
//
// 其余类型返回 ErrNoErrorSample
func BuildErrorSampleSQL(ruleType, db, table, column, paramsJson string) (string, []any, error) {
	switch ruleType {
	case RuleTypeNotNull, RuleTypeUnique, RuleTypeValueRange:
		// 三类均需拼目标列，继续走通用的标识符校验
	default:
		return "", nil, ErrNoErrorSample
	}
	target, err := QualifyTable(db, table)
	if err != nil {
		return "", nil, err
	}
	if !ValidIdentifier(column) {
		return "", nil, fmt.Errorf("%w: target_column %q", ErrInvalidIdentifier, column)
	}

	switch ruleType {
	case RuleTypeNotNull:
		// null 或空串均属脏数据，与检查 SQL 口径一致
		return fmt.Sprintf("select * from %s where `%s` is null or `%s` = '' limit %d",
			target, column, column, errorSampleLimit), nil, nil
	case RuleTypeUnique:
		// join 重复组：取重复值的全部行（t/dup 为固定别名字面量）
		return fmt.Sprintf("select t.* from %s t join "+
			"(select `%s` from %s group by `%s` having count(*) > 1) dup "+
			"on t.`%s` = dup.`%s` limit %d",
			target, column, target, column, column, column, errorSampleLimit), nil, nil
	default: // value_range（入口 switch 已收敛到三类）
		p, err := parseRuleParams(paramsJson)
		if err != nil {
			return "", nil, err
		}
		if !p.hasMin || !p.hasMax {
			return "", nil, fmt.Errorf("%w: value_range 需要 params_json 提供数字 min/max", ErrInvalidRuleParams)
		}
		return fmt.Sprintf("select * from %s where `%s` not between ? and ? limit %d",
			target, column, errorSampleLimit), []any{p.Min, p.Max}, nil
	}
}

// queryErrorRows 用 RawDB 查询任意形状的违规行并转为 []map[string]any
// （select * 的列集合因表而异，go-zero 结构体扫描不适用，走 database/sql 通用扫描：
// 字节列转字符串便于 json 序列化，NULL 存 nil；行数由 SQL 中的 limit 收敛）
func queryErrorRows(ctx context.Context, conn sqlx.SqlConn, query string, args []any, limit int) ([]map[string]any, error) {
	db, err := conn.RawDB()
	if err != nil {
		return nil, fmt.Errorf("获取底层连接失败: %w", err)
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("错误样本查询失败: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("错误样本列解析失败: %w", err)
	}
	out := make([]map[string]any, 0, limit)
	scan := make([]any, len(cols))
	for rows.Next() {
		raw := make([]any, len(cols))
		for i := range raw {
			scan[i] = &raw[i]
		}
		if err := rows.Scan(scan...); err != nil {
			return nil, fmt.Errorf("错误样本行扫描失败: %w", err)
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			switch v := raw[i].(type) {
			case nil:
				row[c] = nil
			case []byte:
				row[c] = string(v)
			case time.Time:
				row[c] = v.Format(time.DateTime)
			default:
				row[c] = v
			}
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// captureOne 抓取单条失败规则的错误样本并落库（供 RunJob 钩子调用）。
// 规则无样本语义、样本为空时静默跳过（不算失败），其余错误原样返回由调用方记日志
func captureOne(ctx context.Context, conn sqlx.SqlConn, errorModel model.ErrorDataModel,
	jobId int64, dbName string, rule *model.Rule, runAt time.Time) error {
	query, args, err := BuildErrorSampleSQL(rule.RuleType, dbName, rule.TargetTable, rule.TargetColumn, rule.ParamsJson)
	if err != nil {
		if errors.Is(err, ErrNoErrorSample) {
			return nil
		}
		return err
	}
	rows, err := queryErrorRows(ctx, conn, query, args, errorSampleLimit)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil // 无样本可存（如检查与抓样窗口间数据已修复），跳过
	}
	rowsJson := marshalDetail(map[string]any{"rows": rows})
	if rowsJson == "" {
		return errors.New("错误样本序列化失败")
	}
	return errorModel.Insert(ctx, &model.ErrorData{
		RuleId:      rule.Id,
		JobId:       jobId,
		RowsJson:    rowsJson,
		SampleCount: int64(len(rows)),
		RunAt:       runAt.UnixMilli(),
	})
}
