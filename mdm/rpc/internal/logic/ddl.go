package logic

import (
	"fmt"
	"strings"
)

// ddl.go 分发目标表 DDL/写入语句拼装（纯函数）：
// 标识符均经 ValidFieldName 校验并用反引号包裹，杜绝注入

// targetIdentifierMaxLen MySQL 标识符长度上限（表名/库名 64 字符）
const targetIdentifierMaxLen = 64

// buildTargetTable 目标表限定名：targetDb 非空时 `db`.`table`，否则 `table`（走 DSN 默认库）
func buildTargetTable(targetDb, targetTable string) string {
	if targetDb == "" {
		return "`" + targetTable + "`"
	}
	return "`" + targetDb + "`.`" + targetTable + "`"
}

// buildCreateTableDDL 按模型字段定义生成幂等建表 DDL（目标表不存在则自动建表）：
// create table if not exists <table> (biz_code varchar(128) primary key,
// <field> <sqltype>..., created_at datetime(3), unique key uk_biz (biz_code))
func buildCreateTableDDL(qualifiedTable string, fields []FieldDef) string {
	var cols strings.Builder
	cols.WriteString("`biz_code` varchar(128) NOT NULL PRIMARY KEY")
	for _, f := range fields {
		cols.WriteString(", `" + f.Name + "` " + SqlTypeOf(f.Type) + " DEFAULT NULL")
	}
	cols.WriteString(", `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3), UNIQUE KEY `uk_biz` (`biz_code`)")
	return fmt.Sprintf("create table if not exists %s (%s) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",
		qualifiedTable, cols.String())
}

// buildReplaceSQL 生成幂等 upsert 语句与占位符个数：
// replace into <table> (biz_code, fields..., created_at) values (...);
// 主键/唯一键冲突时整行替换，实现重复分发幂等
func buildReplaceSQL(qualifiedTable string, fields []FieldDef) (string, int) {
	var cols strings.Builder
	cols.WriteString("`biz_code`")
	for _, f := range fields {
		cols.WriteString(",`" + f.Name + "`")
	}
	cols.WriteString(",`created_at`")
	n := len(fields) + 2 //biz_code + created_at
	return fmt.Sprintf("replace into %s (%s) values (%s)", qualifiedTable, cols.String(), placeholdersN(n)), n
}

// placeholdersN 生成 n 个 ? 占位符（逗号分隔）
func placeholdersN(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}
