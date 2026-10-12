package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"datacenter/job/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// syncengine.go 内置同步执行引擎（Phase 8：DolphinScheduler 不可用时的轻量库到库同步）。
// 组成：纯函数（静态校验/列映射解析/类型推断/DDL·SQL 拼装）+ DB 能力接口注入（分批主循环
// 可 mock 单测）；流程语义对齐 ops archiver：每批一个独立事务，单批失败即停，
// 已提交批保留，失败位置（批号）由调用方结合 RunStats.Batches 写入日志 msg

// ============ 标识符白名单 ============

// identifierRe 标识符白名单：库/表/列名仅允许字母、数字、下划线，
// 不通过者一律拒绝拼接（杜绝 `users; drop table x` 类注入，参考 ops archiver 同款闸门）
var identifierRe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// ValidIdentifier 标识符（库/表/列名）白名单校验，防 SQL 注入
func ValidIdentifier(name string) bool {
	return identifierRe.MatchString(name)
}

// QualifyTable 生成 `<db>`.`<table>` 反引号全限定名（db 为空时走 DSN 默认库），
// 两侧标识符均先过白名单校验
func QualifyTable(db, table string) (string, error) {
	if !ValidIdentifier(table) {
		return "", fmt.Errorf("%w: 表名 %q", ErrInvalidIdentifier, table)
	}
	if db == "" {
		return "`" + table + "`", nil
	}
	if !ValidIdentifier(db) {
		return "", fmt.Errorf("%w: 库名 %q", ErrInvalidIdentifier, db)
	}
	return "`" + db + "`.`" + table + "`", nil
}

// ============ 源端 SELECT 静态校验 ============

// ValidateSourceSql 源端查询静态校验（纯函数；思路参考 ide ValidateSelectSql 并加强）：
//  1. 剥掉前导注释与前导括号后必须以 select 开头（只读限制，拒绝写入语句与 with CTE 包装）；
//  2. 禁分号（单语句限制，天然拒绝 `select 1; drop table x` 多语句注入，结尾分号同拒）；
//  3. 禁注释符（--、#、/*）：引擎会在语句尾部追加 `limit ?, ?` 翻页子句，
//     尾部行注释会吞掉翻页子句导致整表扫描，故一并拒绝（宁严勿漏）；
//  4. 已知限制：语句自带 limit/order by 时与翻页子句叠加为 MySQL 语法错误，
//     错误原样进运行日志 msg（引擎不做语句改写）。
func ValidateSourceSql(sqlText string) error {
	s, ok := stripLeadingNoise(sqlText)
	if !ok {
		return fmt.Errorf("%w: 存在未闭合的块注释", ErrInvalidSourceSql)
	}
	if s == "" {
		return fmt.Errorf("%w: 不能为空", ErrInvalidSourceSql)
	}
	if len(s) < 6 || !strings.EqualFold(s[:6], "select") {
		return fmt.Errorf("%w: 仅允许 SELECT 查询", ErrInvalidSourceSql)
	}
	// 词边界校验：select 后必须是空白或结尾，杜绝 selectx 之类的标识符冒充
	if tail := s[6:]; tail != "" && !isSpaceByte(tail[0]) {
		return fmt.Errorf("%w: 仅允许 SELECT 查询", ErrInvalidSourceSql)
	}
	if strings.Contains(s, ";") {
		return fmt.Errorf("%w: 仅允许单条语句，禁止分号（拒绝 ; drop 等多语句注入片段）", ErrInvalidSourceSql)
	}
	if strings.Contains(s, "--") || strings.Contains(s, "#") || strings.Contains(s, "/*") {
		return fmt.Errorf("%w: 禁止注释符（--/#/\\*，翻页子句拼在语句尾部，注释符会破坏拼接）", ErrInvalidSourceSql)
	}
	return nil
}

// stripLeadingNoise 循环剥掉语句头部的空白、注释与前导括号，返回剩余语句文本；
// ok=false 表示出现未闭合的块注释。循环剥离支持注释与括号任意交错（参考 ide 同名实现）
func stripLeadingNoise(sqlText string) (string, bool) {
	s := sqlText
	for {
		// 跳过前导空白
		for len(s) > 0 && isSpaceByte(s[0]) {
			s = s[1:]
		}
		switch {
		case strings.HasPrefix(s, "--"), strings.HasPrefix(s, "#"):
			// 行注释：跳到行尾（无行尾则整段只剩注释）
			idx := strings.IndexByte(s, '\n')
			if idx < 0 {
				return "", true
			}
			s = s[idx+1:]
		case strings.HasPrefix(s, "/*"):
			// 块注释：未闭合视为非法
			idx := strings.Index(s, "*/")
			if idx < 0 {
				return "", false
			}
			s = s[idx+2:]
		case strings.HasPrefix(s, "("):
			// 前导括号剥离，支持 (select ...) 包装
			s = s[1:]
		default:
			return s, true
		}
	}
}

// isSpaceByte 判断字节是否为空白字符
func isSpaceByte(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

// ============ 内置引擎任务静态校验 ============

// ValidateSyncTaskV2 内置引擎任务静态校验（CreateSyncTaskV2 与 RunSync 双重把关，纯函数）：
// 名称/mode/source_ds_id 复用 V1 规则（validateBase）；source_sql 必填且过 SELECT 单语句校验；
// target_table 必填且过标识符白名单；mapping_json 允许为空（直通）或 [{src,dst,type}] 数组。
// batch_size 不在此拒绝：使用点统一经 model.NormalizeBatchSize 归一（<=0 默认 1000、上限 10000）。
// mode=cdc 本期按 offline 语义执行：引擎暂无位点/binlog 管理，cdc 任务同样做全量分批快照
func ValidateSyncTaskV2(name, mode string, sourceDsId int64,
	sourceSql, targetTable, mappingJson string) error {
	if err := validateBase(name, mode, sourceDsId); err != nil {
		return err
	}
	if strings.TrimSpace(sourceSql) == "" {
		return ErrSourceSqlRequired
	}
	if err := ValidateSourceSql(sourceSql); err != nil {
		return err
	}
	table := strings.TrimSpace(targetTable)
	if table == "" {
		return ErrTargetTableRequired
	}
	if !ValidIdentifier(table) {
		return fmt.Errorf("%w: 目标表名 %q", ErrInvalidIdentifier, table)
	}
	_, err := ParseMapping(mappingJson)
	return err
}

// ============ 列映射解析 ============

// ColumnMapping 列映射（mapping_json 数组元素：[{"src":"user_id","dst":"uid","type":"int"}]）
type ColumnMapping struct {
	Src  string `json:"src"`  // 源结果集列名
	Dst  string `json:"dst"`  // 目标列名
	Type string `json:"type"` // 目标列类型（string/int/decimal/bool/datetime），空=按源列类型推断
}

// ParseMapping 列映射解析（纯函数）：mapping_json 允许为空（按源结果集列名直通，返回 nil）；
// 非空须为 [{src,dst,type}] 数组：src/dst 必填且过标识符白名单、type 过字段类型白名单
// （大小写不敏感）、dst 不得重复
func ParseMapping(mappingJson string) ([]ColumnMapping, error) {
	s := strings.TrimSpace(mappingJson)
	if s == "" {
		return nil, nil
	}
	var list []ColumnMapping
	if err := json.Unmarshal([]byte(s), &list); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidColumnMapping, err)
	}
	seen := make(map[string]struct{}, len(list))
	for i, m := range list {
		m.Src = strings.TrimSpace(m.Src)
		m.Dst = strings.TrimSpace(m.Dst)
		m.Type = strings.ToLower(strings.TrimSpace(m.Type))
		list[i] = m
		if !ValidIdentifier(m.Src) {
			return nil, fmt.Errorf("%w: 第 %d 项 src %q 非法", ErrInvalidColumnMapping, i+1, m.Src)
		}
		if !ValidIdentifier(m.Dst) {
			return nil, fmt.Errorf("%w: 第 %d 项 dst %q 非法", ErrInvalidColumnMapping, i+1, m.Dst)
		}
		if m.Type != "" && !ValidFieldType(m.Type) {
			return nil, fmt.Errorf("%w: 第 %d 项 type %q 不在白名单(string/int/decimal/bool/datetime)",
				ErrInvalidColumnMapping, i+1, m.Type)
		}
		if _, ok := seen[m.Dst]; ok {
			return nil, fmt.Errorf("%w: 目标列 %q 重复", ErrInvalidColumnMapping, m.Dst)
		}
		seen[m.Dst] = struct{}{}
	}
	return list, nil
}

// ============ 字段类型与类型推断（参考 mdm typemap）============

// 字段类型白名单（mapping_json 每项 type 的合法取值；date 替换为 datetime 以保留时分秒）
const (
	FieldTypeString   = "string"
	FieldTypeInt      = "int"
	FieldTypeDecimal  = "decimal"
	FieldTypeBool     = "bool"
	FieldTypeDatetime = "datetime"
)

// sqlTypes 字段类型 → 物理列类型映射表（自动建表按此落列）
var sqlTypes = map[string]string{
	FieldTypeString:   "varchar(255)",
	FieldTypeInt:      "bigint",
	FieldTypeDecimal:  "decimal(18,4)",
	FieldTypeBool:     "tinyint(1)",
	FieldTypeDatetime: "datetime",
}

// ValidFieldType 字段类型是否在白名单
func ValidFieldType(t string) bool {
	_, ok := sqlTypes[t]
	return ok
}

// SqlTypeOfFieldType 字段类型 → 物理列类型（未知类型返回空串，调用方先经 ValidFieldType 把关）
func SqlTypeOfFieldType(t string) string {
	return sqlTypes[t]
}

// InferSqlType 按源列数据库类型名推断目标列类型（纯函数，mapping.type 缺省时用；
// 类型名取自 database/sql ColumnTypes().DatabaseTypeName()）：
// int 家族→bigint、decimal/numeric→decimal(18,4)、bool→tinyint(1)、日期时间族→datetime、
// text/blob/json→text、其余（varchar/char/enum/set/未知）→varchar(255)
func InferSqlType(dbTypeName string) string {
	t := strings.ToUpper(strings.TrimSpace(dbTypeName))
	switch {
	case strings.Contains(t, "INT") || t == "YEAR": // TINYINT/SMALLINT/MEDIUMINT/BIGINT/INTEGER
		return SqlTypeOfFieldType(FieldTypeInt)
	case strings.Contains(t, "DEC") || strings.Contains(t, "NUMERIC") || t == "FIXED": // DECIMAL/NEWDECIMAL
		return SqlTypeOfFieldType(FieldTypeDecimal)
	case strings.Contains(t, "BOOL"): // BOOLEAN
		return SqlTypeOfFieldType(FieldTypeBool)
	case strings.Contains(t, "DATE") || strings.Contains(t, "TIME"): // DATE/DATETIME/TIMESTAMP
		return SqlTypeOfFieldType(FieldTypeDatetime)
	case strings.Contains(t, "TEXT") || strings.Contains(t, "BLOB") ||
		t == "JSON" || t == "BINARY" || t == "VARBINARY": // 大文本/二进制不设长
		return "text"
	default:
		return SqlTypeOfFieldType(FieldTypeString)
	}
}

// ============ 目标列解析与 DDL/SQL 拼装 ============

// TargetColumn 目标列（列名已过白名单，类型为物理列类型）
type TargetColumn struct {
	Name string // 目标列名
	Type string // 物理列类型（建表 DDL 用）
}

// ResolveColumns 目标列解析（纯函数，首批数据到达后执行一次）：
// mapping 非空 → 按 [src→dst] 投影（src 在结果集中按大小写不敏感匹配，缺列即报错），
// type 缺省按源列数据库类型名推断；mapping 缺省 → 按源结果集列名直通
// （列名必须过白名单，表达式列需在源 SQL 里用 AS 起别名）。
// 返回目标列与各列在源结果集中的下标（写入时按位取值）
func ResolveColumns(resultCols, resultTypes []string, mapping []ColumnMapping) ([]TargetColumn, []int, error) {
	if len(resultCols) == 0 {
		return nil, nil, fmt.Errorf("%w: 源结果集无列", ErrInvalidSourceSql)
	}
	if len(mapping) == 0 {
		// 缺省直通：目标列名=源列名，类型按数据库类型名推断
		cols := make([]TargetColumn, 0, len(resultCols))
		idx := make([]int, 0, len(resultCols))
		for i, c := range resultCols {
			if !ValidIdentifier(c) {
				return nil, nil, fmt.Errorf("%w: 源结果集列名 %q 非法（表达式列请在源 SQL 中用 AS 别名）",
					ErrInvalidIdentifier, c)
			}
			cols = append(cols, TargetColumn{Name: c, Type: InferSqlType(resultTypes[i])})
			idx = append(idx, i)
		}
		return cols, idx, nil
	}
	// 有映射：按 [src→dst] 投影
	lowerIdx := make(map[string]int, len(resultCols)) // 列名小写 → 结果集下标（匹配大小写不敏感）
	for i, c := range resultCols {
		lowerIdx[strings.ToLower(strings.TrimSpace(c))] = i
	}
	cols := make([]TargetColumn, 0, len(mapping))
	idx := make([]int, 0, len(mapping))
	for _, m := range mapping {
		i, ok := lowerIdx[strings.ToLower(m.Src)]
		if !ok {
			return nil, nil, fmt.Errorf("%w: 源结果集缺少映射列 %q", ErrInvalidColumnMapping, m.Src)
		}
		typ := m.Type
		if typ == "" {
			typ = InferSqlType(resultTypes[i])
		} else {
			// 字段类型白名单值 → 物理列类型（string→varchar(255)、decimal→decimal(18,4) 等）
			typ = SqlTypeOfFieldType(typ)
		}
		cols = append(cols, TargetColumn{Name: m.Dst, Type: typ})
		idx = append(idx, i)
	}
	return cols, idx, nil
}

// buildCreateTargetDDL 目标表不存在时自动建表（幂等 DDL）：
// create table if not exists `<db>`.`<t>` (`id` bigint unsigned AUTO_INCREMENT 主键,
// `col` type DEFAULT NULL, ...)。自增 id 仅作物理主键；replace 的幂等键由目标表上的
// 业务唯一键决定（如需按业务键幂等可自建唯一索引，replace 冲突时整行替换）
func buildCreateTargetDDL(qualifiedTable string, cols []TargetColumn) string {
	var b strings.Builder
	b.WriteString("create table if not exists " + qualifiedTable +
		" (`id` bigint unsigned NOT NULL AUTO_INCREMENT")
	for _, c := range cols {
		b.WriteString(", `" + c.Name + "` " + c.Type + " DEFAULT NULL")
	}
	b.WriteString(", PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	return b.String()
}

// buildReplaceSQL 单批 replace into 语句与占位符总数（多行 values 一次写入）：
// replace into `<db>`.`<t>` (`c1`,`c2`) values (?,?),(?,?)...
func buildReplaceSQL(qualifiedTable string, cols []TargetColumn, rowCount int) (string, int) {
	var b strings.Builder
	b.WriteString("replace into " + qualifiedTable + " (")
	for i, c := range cols {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("`" + c.Name + "`")
	}
	b.WriteString(") values ")
	rowPh := "(" + placeholdersN(len(cols)) + ")"
	for i := 0; i < rowCount; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(rowPh)
	}
	return b.String(), len(cols) * rowCount
}

// placeholdersN 生成 n 个 ? 占位符（逗号分隔）
func placeholdersN(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// projectRows 按 src 下标投影行数据（idx 为空时原样直通）
func projectRows(rows [][]any, idx []int) [][]any {
	if len(idx) == 0 {
		return rows
	}
	out := make([][]any, 0, len(rows))
	for _, r := range rows {
		row := make([]any, 0, len(idx))
		for _, i := range idx {
			row = append(row, r[i])
		}
		out = append(out, row)
	}
	return out
}

// ============ 分批主循环（DB 能力注入）============

// BatchData 源结果集单批数据（RawDB 手工扫描：列名 + 驱动原值）
type BatchData struct {
	Columns []string // 结果集列名
	Types   []string // 各列数据库类型名（ColumnTypes().DatabaseTypeName()，缺省类型推断用）
	Rows    [][]any  // 行数据（驱动原值，按列序）
}

// SyncSourceDB 源库读取能力抽象（接口注入，RunSyncLoop 可 mock 单测）
type SyncSourceDB interface {
	// ReadBatch 读取源查询的第 offset 批（limit 行）
	ReadBatch(ctx context.Context, sourceSql string, offset, limit int64) (*BatchData, error)
}

// SyncTargetDB 目标库写入能力抽象
type SyncTargetDB interface {
	// EnsureTable 目标表不存在则按列定义自动建表（幂等）
	EnsureTable(ctx context.Context, qualifiedTable string, cols []TargetColumn) error
	// WriteBatch 单批事务写入（replace into 多行 values），返回写入行数；单批失败整批回滚
	WriteBatch(ctx context.Context, qualifiedTable string, cols []TargetColumn, rows [][]any) (int64, error)
}

// MaxSyncBatches 单次同步最大批数安全上限（10000 批 × 10000 行兜底，防御 offset 翻页
// 异常导致的死循环，与 ops MaxArchiveBatches 同语义）
const MaxSyncBatches = 10000

// RunStats 一次内置同步的累计结果
type RunStats struct {
	ReadRows  int64 // 累计读取行数（含失败前已读批）
	WriteRows int64 // 累计写入行数（失败中断时为已提交批）
	Batches   int64 // 已提交事务批数（失败中断时为失败前成功提交的批数，失败批号=此值+1）
}

// SyncOptions 一次内置同步的输入（SQL/标识符均已过校验，表名为全限定名）
type SyncOptions struct {
	SourceSql      string          // 源端 SELECT（已过 ValidateSourceSql）
	QualifiedTable string          // 目标表全限定名（已过 QualifyTable）
	Mapping        []ColumnMapping // 列映射（nil=按源结果集列名直通）
	BatchSize      int64           // 分批大小（<=0 时取默认 1000）
}

// RunSyncLoop 内置同步分批主循环（纯流程编排，DB 能力注入可 mock 单测）：
// 源 `source_sql limit ?, ?` offset 翻页读 → 首批解析目标列并自动建表 →
// 目标 replace into 分批写（每批一个独立事务）→ 累计 read/write；
// 批行数不足 batch_size 或读空即正常收尾；任一批失败立即中断（已提交批保留），
// 错误原样上抛，由调用方结合 RunStats.Batches 拼装"第 N 批失败"msg
func RunSyncLoop(ctx context.Context, src SyncSourceDB, tgt SyncTargetDB, opt SyncOptions) (RunStats, error) {
	batchSize := model.NormalizeBatchSize(opt.BatchSize)
	var stats RunStats
	var cols []TargetColumn // 首批确定后固定
	var idx []int
	for stats.Batches < MaxSyncBatches {
		data, err := src.ReadBatch(ctx, opt.SourceSql, stats.Batches*batchSize, batchSize)
		if err != nil {
			return stats, fmt.Errorf("读取源数据失败: %w", err)
		}
		// 首批确定目标列并确保目标表存在；源无数据时无从取列，跳过建表直接成功收尾
		if cols == nil {
			if len(data.Rows) == 0 {
				return stats, nil
			}
			if cols, idx, err = ResolveColumns(data.Columns, data.Types, opt.Mapping); err != nil {
				return stats, err
			}
			if err = tgt.EnsureTable(ctx, opt.QualifiedTable, cols); err != nil {
				return stats, fmt.Errorf("目标表 %s 建表失败: %w", opt.QualifiedTable, err)
			}
		}
		if len(data.Rows) == 0 { // offset 翻页读空：正常收尾
			return stats, nil
		}
		n, err := tgt.WriteBatch(ctx, opt.QualifiedTable, cols, projectRows(data.Rows, idx))
		if err != nil { // 单批失败即停：已提交批保留
			return stats, fmt.Errorf("写入目标表失败: %w", err)
		}
		stats.ReadRows += int64(len(data.Rows))
		stats.WriteRows += n
		stats.Batches++
		if int64(len(data.Rows)) < batchSize { // 不足一批：已到最后一批
			return stats, nil
		}
	}
	// 达到批数上限：强制收尾兜底（防死循环），按完成口径返回
	return stats, nil
}

// NextTaskStatus 内置同步完成后的任务状态流转（纯函数）：
// 运行失败 → failed；运行成功 → 草稿保持草稿（内置引擎直跑不代表已部署 DS）、
// 其余状态回落 deployed 终态（与 DS 触发链路的状态机一致）
func NextTaskStatus(local, runStatus string) string {
	if runStatus == model.SyncLogStatusFailed {
		return model.TaskStatusFailed
	}
	if local == model.TaskStatusDraft {
		return model.TaskStatusDraft
	}
	return model.TaskStatusDeployed
}

// ============ mysql 实现 ============

// mysqlSourceDB SyncSourceDB 的 mysql 实现（RawDB 手工扫描，参考 tag taggingLogic：
// 列数不定无法落固定 struct，按容器指针逐列扫描；驱动 []byte/time.Time 原值透传到目标库）
type mysqlSourceDB struct {
	db *sql.DB
}

// NewMysqlSourceDB 构造源端读取器（RawDB 复用 SqlConn 底层连接池，不可 Close）
func NewMysqlSourceDB(conn sqlx.SqlConn) (SyncSourceDB, error) {
	db, err := conn.RawDB()
	if err != nil {
		return nil, err
	}
	return &mysqlSourceDB{db: db}, nil
}

// ReadBatch 执行 `source_sql limit ?, ?` 并整批读入内存（batch_size 上限 10000，内存可控）。
// sourceSql 已过 ValidateSourceSql（单条 SELECT、无分号/注释），尾部拼接翻页子句安全
func (m *mysqlSourceDB) ReadBatch(ctx context.Context, sourceSql string, offset, limit int64) (*BatchData, error) {
	rs, err := m.db.QueryContext(ctx, sourceSql+" limit ?, ?", offset, limit)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	cols, err := rs.Columns()
	if err != nil {
		return nil, err
	}
	// 数据库类型名仅服务类型推断，获取失败不致命（推断回落 varchar(255)）
	types := make([]string, len(cols))
	if cts, terr := rs.ColumnTypes(); terr == nil {
		for i, ct := range cts {
			types[i] = ct.DatabaseTypeName()
		}
	}
	data := &BatchData{Columns: cols, Types: types}
	for rs.Next() {
		container := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range container {
			ptrs[i] = &container[i]
		}
		if err := rs.Scan(ptrs...); err != nil {
			return nil, err
		}
		data.Rows = append(data.Rows, container)
	}
	return data, rs.Err()
}

// mysqlTargetDB SyncTargetDB 的 mysql 实现
type mysqlTargetDB struct {
	conn sqlx.SqlConn
}

// NewMysqlTargetDB 构造目标写入器
func NewMysqlTargetDB(conn sqlx.SqlConn) SyncTargetDB {
	return &mysqlTargetDB{conn: conn}
}

// EnsureTable 目标表不存在则自动建表（幂等 DDL；已存在时不动，缺列由写入报错暴露）
func (m *mysqlTargetDB) EnsureTable(ctx context.Context, qualifiedTable string, cols []TargetColumn) error {
	_, err := m.conn.ExecCtx(ctx, buildCreateTargetDDL(qualifiedTable, cols))
	return err
}

// WriteBatch 单批一个事务：replace into 多行 values 一次提交，失败整批回滚（已提交批不受影响）。
// 写入行数取 len(rows) 而非 RowsAffected：replace 命中唯一键时按"删除+插入"计 2，计数虚高
func (m *mysqlTargetDB) WriteBatch(ctx context.Context, qualifiedTable string, cols []TargetColumn, rows [][]any) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	query, phTotal := buildReplaceSQL(qualifiedTable, cols, len(rows))
	args := make([]any, 0, phTotal)
	for _, r := range rows {
		args = append(args, r...)
	}
	if err := m.conn.TransactCtx(ctx, func(tx context.Context, session sqlx.Session) error {
		_, err := session.ExecCtx(tx, query, args...)
		return err
	}); err != nil {
		return 0, err
	}
	return int64(len(rows)), nil
}

// ============ 运行日志 msg 工具 ============

// syncMsgMaxLen 运行日志 msg 列宽上限（dc_sync_task_log.msg varchar(1024)，按字符计）
const syncMsgMaxLen = 1024

// truncateRunes 按字符数截断（MySQL varchar(n) 按字符计，参考 ops convert 同名实现）
func truncateRunes(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n])
}
