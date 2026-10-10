package logic

import (
	"context"
	"database/sql"
)

// QueryRowsAsMaps 在目标库执行 SELECT，把结果集转为通用行集（手工扫描模式，参考 datasvc executor.go）：
//   - 按结果集列名定位取值，不依赖扫描 struct，规避 go-zero sqlx 少列查询扫宽 struct
//     的 "not matching destination to scan" 报错；
//   - 驱动返回的 []byte（varchar/text/decimal 等）统一转 string，便于 JSON 输出；
//   - maxRows 保护性截断：达到行数上限且仍有下一行时停止拉取，truncated=true 标识；
//     提前 break 后由 defer rs.Close() 归还连接，剩余行由驱动丢弃。
func QueryRowsAsMaps(ctx context.Context, db *sql.DB, query string, maxRows int) ([]string, []map[string]interface{}, bool, error) {
	rs, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, false, err
	}
	defer rs.Close()
	cols, err := rs.Columns()
	if err != nil {
		return nil, nil, false, err
	}
	if cols == nil {
		cols = []string{}
	}
	truncated := false
	rows := make([]map[string]interface{}, 0, 16)
	for rs.Next() {
		if len(rows) >= maxRows {
			//已达上限且仍有下一行 → 截断
			truncated = true
			break
		}
		container := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range container {
			ptrs[i] = &container[i]
		}
		if err := rs.Scan(ptrs...); err != nil {
			return nil, nil, false, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, c := range cols {
			row[c] = normalizeDbValue(container[i])
		}
		rows = append(rows, row)
	}
	if err := rs.Err(); err != nil {
		return nil, nil, false, err
	}
	return cols, rows, truncated, nil
}

// normalizeDbValue 驱动返回值归一化（仅 []byte → string，其余保持原样）
func normalizeDbValue(v interface{}) interface{} {
	switch t := v.(type) {
	case []byte:
		return string(t)
	default:
		return v
	}
}
