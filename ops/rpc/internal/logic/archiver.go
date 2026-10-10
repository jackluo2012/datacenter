package logic

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ============ 标识符白名单 ============

// identifierRe 标识符白名单：表名/库名仅允许字母、数字、下划线，
// 不通过者一律拒绝拼接（杜绝 `users; drop table x` 类注入）
var identifierRe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// ValidIdentifier 标识符（库/表名）白名单校验，防 SQL 注入
func ValidIdentifier(name string) bool {
	return identifierRe.MatchString(name)
}

// QualifyTable 生成 `<db>.<table>` 全限定表名（db 为空时使用 DSN 默认库），
// 两侧标识符均先过白名单校验
func QualifyTable(db, table string) (string, error) {
	if !ValidIdentifier(table) {
		return "", fmt.Errorf("%w: 表名 %q", ErrInvalidIdentifier, table)
	}
	if db == "" {
		return table, nil
	}
	if !ValidIdentifier(db) {
		return "", fmt.Errorf("%w: 库名 %q", ErrInvalidIdentifier, db)
	}
	return db + "." + table, nil
}

// ============ 归档条件（condition_sql）静态校验 ============

// maxConditionLength condition_sql 列宽上限（varchar(1024)，按字符计）
const maxConditionLength = 1024

// condForbiddenWords 归档条件黑名单关键字（词边界全词匹配，避免误伤 created_at 之类列名）。
// 条件片段只允许纯 where 过滤（可含子查询），凡能改变语句形态或执行危险动作的词一律拒绝
var condForbiddenWords = []string{
	"drop", "truncate", "alter", "create", "rename", "grant", "revoke",
	"insert", "update", "delete", "union", "outfile", "load_file",
	"benchmark", "sleep", "shutdown", "procedure", "handler", "information_schema",
}

// condForbiddenWordRes 预编译的黑名单词边界正则（与 condForbiddenWords 下标一一对应）
var condForbiddenWordRes = func() []*regexp.Regexp {
	res := make([]*regexp.Regexp, 0, len(condForbiddenWords))
	for _, w := range condForbiddenWords {
		res = append(res, regexp.MustCompile(`(?i)\b`+w+`\b`))
	}
	return res
}()

// ValidateConditionSQL 归档条件静态校验（创建任务与运行前双重把关）：
//  1. 允许为空（整表归档）
//  2. 非空必须以 where 开头（只接受过滤片段）
//  3. 剔除引号字面量后做黑名单校验：禁分号、禁 --/#//*
//     注释、禁 drop/truncate/alter 等危险关键字全词出现
func ValidateConditionSQL(cond string) error {
	trimmed := strings.TrimSpace(cond)
	if trimmed == "" {
		return nil
	}
	if len([]rune(trimmed)) > maxConditionLength {
		return ErrConditionTooLong
	}
	// 前缀必须是 where（大小写不敏感）
	if len(trimmed) < 5 || !strings.EqualFold(trimmed[:5], "where") {
		return fmt.Errorf("%w: 须以 where 开头", ErrInvalidCondition)
	}
	// 引号字面量先剔除再查黑名单：'a;b'、'DROP' 等字面量内容不构成注入面
	stripped, err := stripQuotedLiterals(trimmed)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCondition, err)
	}
	// 字符级黑名单：分号（多语句）、行注释与块注释
	if strings.Contains(stripped, ";") {
		return fmt.Errorf("%w: 禁止分号", ErrInvalidCondition)
	}
	if strings.Contains(stripped, "--") || strings.Contains(stripped, "#") || strings.Contains(stripped, "/*") {
		return fmt.Errorf("%w: 禁止注释符（--/#/\\*）", ErrInvalidCondition)
	}
	// 关键字黑名单（全词边界匹配）
	for i, re := range condForbiddenWordRes {
		if re.MatchString(stripped) {
			return fmt.Errorf("%w: 禁止出现关键字 %q", ErrInvalidCondition, condForbiddenWords[i])
		}
	}
	return nil
}

// stripQuotedLiterals 把 SQL 中的单/双引号字面量整体原位替换为空格，
// 支持反斜杠转义（\'）与双写转义（两个连续引号）；引号未闭合视为非法
func stripQuotedLiterals(s string) (string, error) {
	out := []byte(s)
	i := 0
	for i < len(out) {
		c := out[i]
		if c != '\'' && c != '"' {
			i++
			continue
		}
		quote := c
		out[i] = ' '
		i++
		closed := false
		for i < len(out) && !closed {
			switch {
			case out[i] == '\\' && i+1 < len(out): // 反斜杠转义（MySQL 字符串语法）
				out[i], out[i+1] = ' ', ' '
				i += 2
			case out[i] == quote && i+1 < len(out) && out[i+1] == quote: // 双写转义（''/""）
				out[i], out[i+1] = ' ', ' '
				i += 2
			case out[i] == quote: // 字面量正常闭合
				out[i] = ' '
				i++
				closed = true
			default: // 字面量内容，原位抹成空格
				out[i] = ' '
				i++
			}
		}
		if !closed {
			return "", errors.New("引号字面量未闭合")
		}
	}
	return string(out), nil
}

// ============ 分批归档执行 ============

// DefaultBatchSize 批大小缺省值（batch_size<=0 时兜底）
const DefaultBatchSize = 1000

// MaxArchiveBatches 单次执行的最大批数安全上限。
// delete_source=true 时依赖"删源推进"，正常按 0 行自然终止；
// 该上限兜底防御异常条件/数据导致的死循环（10000 批 * 1000 行为千万行量级）
const MaxArchiveBatches = 10000

// BatchSpec 单批归档执行参数（表名为已过白名单的全限定名，条件为已过黑名单的 where 片段）
type BatchSpec struct {
	SourceTable  string // 源表
	ArchiveTable string // 归档表
	Condition    string // where 条件片段（可为空）
	Offset       int64  // 翻页偏移，仅 delete_source=false 的纯复制模式使用；-1 表示不带 offset
	Limit        int64  // 本批行数
	DeleteSource bool   // 是否同事务删除源表对应行
}

// BatchResult 单批执行结果
type BatchResult struct {
	Archived int64 // 本批插入归档表行数
	Deleted  int64 // 本批删除源表行数
}

// ArchiveDB 归档目标库操作接口：DB 能力抽象成接口注入，
// RunArchiveLoop 可注入 mock 单测累计/停止/失败中断逻辑，生产走 mysql 实现
type ArchiveDB interface {
	// ArchiveBatch 执行一批归档；DeleteSource=true 时 insert 与 delete 在同一事务内提交，
	// 保证"已归档必已删源、已删源必已归档"，不丢数据也不重复
	ArchiveBatch(ctx context.Context, spec BatchSpec) (BatchResult, error)
}

// RunStats 归档累计结果
type RunStats struct {
	Archived int64 // 累计归档行数
	Deleted  int64 // 累计删除源表行数
	Batches  int64 // 已执行批数（失败中断时为失败前已成功提交的批数）
}

// ArchiveOptions 一次归档执行的选项
type ArchiveOptions struct {
	SourceTable  string
	ArchiveTable string
	Condition    string
	BatchSize    int64
	DeleteSource bool
	MaxBatches   int64 // <=0 时取 MaxArchiveBatches
}

// RunArchiveLoop 分批归档主循环（纯流程编排，DB 能力由 ArchiveDB 注入）：
// 逐批执行，单批影响行数为 0 即正常收尾；任一批失败立即中断（已提交批保留），
// 错误原样上抛，由调用方结合 RunStats.Batches 定位失败批号
func RunArchiveLoop(ctx context.Context, db ArchiveDB, opt ArchiveOptions) (RunStats, error) {
	if opt.BatchSize <= 0 {
		opt.BatchSize = DefaultBatchSize
	}
	if opt.MaxBatches <= 0 {
		opt.MaxBatches = MaxArchiveBatches
	}
	var stats RunStats
	// delete 模式每批删源推进，恒从头部取，不带 offset；
	// 纯复制模式（仅插入）源表数据不动，用 offset 翻页，避免每批反复归档同一批行
	offset := int64(-1)
	if !opt.DeleteSource {
		offset = 0
	}
	for stats.Batches < opt.MaxBatches {
		res, err := db.ArchiveBatch(ctx, BatchSpec{
			SourceTable:  opt.SourceTable,
			ArchiveTable: opt.ArchiveTable,
			Condition:    opt.Condition,
			Offset:       offset,
			Limit:        opt.BatchSize,
			DeleteSource: opt.DeleteSource,
		})
		if err != nil { // 单批失败即停：已提交批保留
			return stats, err
		}
		stats.Archived += res.Archived
		stats.Deleted += res.Deleted
		stats.Batches++
		if res.Archived == 0 { // 影响行数 0：已无待归档数据
			break
		}
		if offset >= 0 {
			offset += res.Archived
		}
	}
	return stats, nil
}

// mysqlArchiveDB ArchiveDB 的 mysql 实现（目标库连接由数据源服务 DSN 构造）
type mysqlArchiveDB struct {
	conn sqlx.SqlConn
}

// NewMysqlArchiveDB 构造 mysql 归档执行器
func NewMysqlArchiveDB(conn sqlx.SqlConn) ArchiveDB {
	return &mysqlArchiveDB{conn: conn}
}

// ArchiveBatch 单批归档：insert into <archive> select * from <src> <cond> limit <n>，
// DeleteSource=true 时同事务内再 delete from <src> <cond> limit <n>。
// insert 未命中（0 行）时不动源表，避免误删尚未归档的行；
// insert 成功而 delete 失败时整个事务回滚，不会出现"删了源没归档"。
// 说明：insert/delete 采用同一条件片段同序扫描，MySQL 常规规划下取到同一批行；
// 若源表存在高并发写入，两语句取行可能漂移，归档窗口期建议停写或接受极小概率偏差
func (m *mysqlArchiveDB) ArchiveBatch(ctx context.Context, spec BatchSpec) (BatchResult, error) {
	var res BatchResult
	insSQL, insArgs := buildInsertSQL(spec)
	delSQL, delArgs := buildDeleteSQL(spec)
	err := m.conn.TransactCtx(ctx, func(tx context.Context, session sqlx.Session) error {
		ret, err := session.ExecCtx(tx, insSQL, insArgs...)
		if err != nil {
			return err
		}
		n, err := ret.RowsAffected()
		if err != nil {
			return err
		}
		res.Archived = n
		// 无行可归档时不删源
		if n == 0 || !spec.DeleteSource {
			return nil
		}
		ret, err = session.ExecCtx(tx, delSQL, delArgs...)
		if err != nil { // 返回 error 触发事务回滚，本批 insert 一并撤销
			return err
		}
		res.Deleted, err = ret.RowsAffected()
		return err
	})
	if err != nil {
		return BatchResult{}, err
	}
	return res, nil
}

// buildInsertSQL 生成单批插入 SQL 与参数（表名已过白名单、条件已过黑名单，此处直接拼接）：
//
//	常规：insert into <archive> select * from <src> <cond> limit ?
//	纯复制模式带 offset 翻页：insert into <archive> select * from <src> <cond> limit ?, ?
func buildInsertSQL(spec BatchSpec) (string, []any) {
	query := fmt.Sprintf("insert into %s select * from %s", spec.ArchiveTable, spec.SourceTable)
	if spec.Condition != "" {
		query += " " + spec.Condition
	}
	var args []any
	if spec.Offset >= 0 {
		query += " limit ?, ?"
		args = append(args, spec.Offset, spec.Limit)
	} else {
		query += " limit ?"
		args = append(args, spec.Limit)
	}
	return query, args
}

// buildDeleteSQL 生成单批删除源表 SQL 与参数：delete from <src> <cond> limit ?
func buildDeleteSQL(spec BatchSpec) (string, []any) {
	query := fmt.Sprintf("delete from %s", spec.SourceTable)
	if spec.Condition != "" {
		query += " " + spec.Condition
	}
	query += " limit ?"
	return query, []any{spec.Limit}
}
