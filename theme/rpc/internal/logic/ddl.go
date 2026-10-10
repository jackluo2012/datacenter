package logic

import (
	"fmt"
	"regexp"
	"strings"
)

// ddl.go 建表/建视图 DDL 拼装（纯函数）：
// 表名/视图名/库名/列名均先经白名单正则校验再用反引号包裹，杜绝注入

// qualifiedIdentRe 限定名白名单（进 DDL 前的最后一道闸）：剥去反引号与点号后，
// 必须仅含字母/数字/下划线（^[A-Za-z0-9_]+$），否则拒绝拼接
var qualifiedIdentRe = regexp.MustCompile(`^[A-Za-z0-9_.]+$`)

// validQualifiedIdent 限定名（`db`.`table` 或 `table`）是否通过白名单：
// 剥去反引号后必须仅含字母/数字/下划线/点号
func validQualifiedIdent(s string) bool {
	return qualifiedIdentRe.MatchString(strings.ReplaceAll(s, "`", ""))
}

// buildPhysicalTable 分层物理表名：<layer>_<table_name>（如 dwd_order_detail）
func buildPhysicalTable(layer, tableName string) string {
	return layer + "_" + tableName
}

// buildQualifiedTable 限定名：db 非空时 `db`.`table`，否则 `table`（走 DSN 默认库）
func buildQualifiedTable(db, table string) string {
	if db == "" {
		return "`" + table + "`"
	}
	return "`" + db + "`.`" + table + "`"
}

// buildCreateTableDDL 按字段定义生成分层建表 DDL（幂等，表不存在则自动建）：
// create table if not exists <layer>_<table> (`f1` type, ..., primary key(`pk`))
//   - 主键（pk=true）必须恰有一个，由 pk 断言把守；
//   - 主键列落 NOT NULL，其余列 DEFAULT NULL；
//   - 列名/限定名任一未过白名单即拒绝（防注入——本 DDL 会真实执行）。
func buildCreateTableDDL(qualifiedTable string, fields []FieldDef) (string, error) {
	if !validQualifiedIdent(qualifiedTable) {
		return "", ErrInvalidTableName
	}
	if len(fields) == 0 {
		return "", fmt.Errorf("%w: 字段定义不能为空", ErrInvalidFieldsJson)
	}
	pk := ""
	pkCount := 0
	cols := make([]string, 0, len(fields))
	for _, f := range fields {
		if !ValidFieldName(f.Name) {
			return "", fmt.Errorf("%w: 字段名 %s 不合法（须以字母开头，仅含字母/数字/下划线）",
				ErrInvalidFieldsJson, f.Name)
		}
		col := "`" + f.Name + "` " + SqlTypeOf(f.Type)
		if f.Pk {
			col += " NOT NULL"
			pk = f.Name
			pkCount++
		} else {
			col += " DEFAULT NULL"
		}
		cols = append(cols, col)
	}
	//pk 断言：主键必须恰有一个（部署前存量数据防御，保存时已把过一轮）
	if pkCount != 1 {
		return "", fmt.Errorf("%w: 主键字段（pk）必须恰有一个，当前 %d 个", ErrInvalidFieldsJson, pkCount)
	}
	return fmt.Sprintf("create table if not exists %s (%s, primary key(`%s`)) "+
		"ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",
		qualifiedTable, strings.Join(cols, ", "), pk), nil
}

// buildCreateViewDDL 物化视图刷新 DDL（MySQL 普通视图即可满足"物化"语义，
// create or replace view 原子替换，等效 drop+create 的重建效果）：
// create or replace view <view> as <select_sql>
func buildCreateViewDDL(qualifiedView, selectSql string) string {
	return "create or replace view " + qualifiedView + " as " + strings.TrimSpace(selectSql)
}

// summarizeSql SQL 摘要（供返回 msg 携带生成 SQL）：超长截断加省略号
func summarizeSql(sql string, maxLen int) string {
	s := strings.TrimSpace(sql)
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
