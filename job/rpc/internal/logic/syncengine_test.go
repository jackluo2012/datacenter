package logic

import (
	"context"
	"errors"
	"strings"
	"testing"

	"datacenter/job/model"
)

// syncengine 单测（标准库）：静态校验（注入拒绝/白名单）、列映射解析、类型推断、
// 分批主循环（mock 注入：累计/中断/已提交批保留）、DDL/SQL 拼装

// TestValidateSyncTaskV2 内置引擎任务静态校验：注入拒绝、白名单、必填项
func TestValidateSyncTaskV2(t *testing.T) {
	legalMapping := `[{"src":"user_id","dst":"uid","type":"int"}]`
	cases := []struct {
		name        string
		taskName    string
		mode        string
		sourceDsId  int64
		sourceSql   string
		targetTable string
		mappingJson string
		wantErr     error
	}{
		{name: "离线任务合法", taskName: "订单同步", mode: "offline", sourceDsId: 1,
			sourceSql: "select id,user_name from orders", targetTable: "dwd_orders"},
		{name: "cdc 任务合法(本期按 offline 语义执行)", taskName: "订单同步", mode: "cdc", sourceDsId: 1,
			sourceSql: "select id from orders", targetTable: "t1", mappingJson: legalMapping},
		{name: "带前导注释与括号的 select 合法", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "/* 注释 */ (select id from t)", targetTable: "t1"},
		{name: "名称为空", taskName: "  ", mode: "offline", sourceDsId: 1,
			sourceSql: "select 1", targetTable: "t1", wantErr: ErrTaskNameRequired},
		{name: "mode 不在白名单", taskName: "t", mode: "stream", sourceDsId: 1,
			sourceSql: "select 1", targetTable: "t1", wantErr: ErrInvalidMode},
		{name: "源数据源缺失", taskName: "t", mode: "offline", sourceDsId: 0,
			sourceSql: "select 1", targetTable: "t1", wantErr: ErrSourceDsRequired},
		{name: "source_sql 缺失", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "  ", targetTable: "t1", wantErr: ErrSourceSqlRequired},
		{name: "source_sql 多语句注入拒绝", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "select 1; drop table x", targetTable: "t1", wantErr: ErrInvalidSourceSql},
		{name: "source_sql 写语句拒绝", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "update t set a=1", targetTable: "t1", wantErr: ErrInvalidSourceSql},
		{name: "source_sql with CTE 拒绝", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "with c as (select 1) select * from c", targetTable: "t1", wantErr: ErrInvalidSourceSql},
		{name: "source_sql selectx 冒充拒绝", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "selectx 1", targetTable: "t1", wantErr: ErrInvalidSourceSql},
		{name: "source_sql 未闭合块注释拒绝", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "/* comment", targetTable: "t1", wantErr: ErrInvalidSourceSql},
		{name: "source_sql 尾部行注释拒绝(会吞翻页子句)", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "select 1 -- note", targetTable: "t1", wantErr: ErrInvalidSourceSql},
		{name: "source_sql 井号注释拒绝", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "select 1 # note", targetTable: "t1", wantErr: ErrInvalidSourceSql},
		{name: "target_table 缺失", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "select 1", targetTable: "  ", wantErr: ErrTargetTableRequired},
		{name: "target_table 注入拒绝", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "select 1", targetTable: "t; drop table x", wantErr: ErrInvalidIdentifier},
		{name: "target_table 含连字符拒绝", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "select 1", targetTable: "t-1", wantErr: ErrInvalidIdentifier},
		{name: "mapping_json 非 JSON 拒绝", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "select 1", targetTable: "t1", mappingJson: "{bad", wantErr: ErrInvalidColumnMapping},
		{name: "mapping_json 须为数组(V1 对象语义不适用)", taskName: "t", mode: "offline", sourceDsId: 1,
			sourceSql: "select 1", targetTable: "t1", mappingJson: `{"table":"x"}`, wantErr: ErrInvalidColumnMapping},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateSyncTaskV2(c.taskName, c.mode, c.sourceDsId,
				c.sourceSql, c.targetTable, c.mappingJson)
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("ValidateSyncTaskV2 = %v, want %v", err, c.wantErr)
			}
		})
	}
}

// TestValidateSourceSql 源查询校验细节
func TestValidateSourceSql(t *testing.T) {
	if err := ValidateSourceSql("  select id , name from t where a = 'x'  "); err != nil {
		t.Fatalf("合法 select 不应报错: %v", err)
	}
	if err := ValidateSourceSql("select 1; "); err == nil {
		t.Fatalf("结尾分号应拒绝")
	}
	// 字面量内的分号同样拒绝（宁严勿漏，与 ide 口径一致）
	if err := ValidateSourceSql("select 'a;b' from t"); err == nil {
		t.Fatalf("字面量内分号应拒绝")
	}
	if err := ValidateSourceSql(""); err == nil {
		t.Fatalf("空 SQL 应拒绝")
	}
}

// TestParseMapping 列映射解析
func TestParseMapping(t *testing.T) {
	// 空 = 缺省直通
	m, err := ParseMapping("  ")
	if err != nil || m != nil {
		t.Fatalf("空 mapping 应返回 nil,nil, got %v,%v", m, err)
	}
	// 合法数组（type 大小写不敏感归一）
	m, err = ParseMapping(`[{"src":"ID","dst":"uid","type":" INT "},{"src":"name","dst":"uname"}]`)
	if err != nil {
		t.Fatalf("合法 mapping 不应报错: %v", err)
	}
	if len(m) != 2 || m[0].Type != "int" || m[1].Type != "" {
		t.Fatalf("mapping 归一结果不符: %+v", m)
	}
	// 非法 JSON
	if _, err := ParseMapping("[{bad}"); !errors.Is(err, ErrInvalidColumnMapping) {
		t.Fatalf("非法 JSON 应返回 ErrInvalidColumnMapping, got %v", err)
	}
	// dst 注入拒绝
	if _, err := ParseMapping(`[{"src":"a","dst":"b; drop table x"}]`); !errors.Is(err, ErrInvalidColumnMapping) {
		t.Fatalf("dst 注入应拒绝, got %v", err)
	}
	// type 不在白名单
	if _, err := ParseMapping(`[{"src":"a","dst":"b","type":"float"}]`); !errors.Is(err, ErrInvalidColumnMapping) {
		t.Fatalf("type 越界应拒绝, got %v", err)
	}
	// dst 重复
	if _, err := ParseMapping(`[{"src":"a","dst":"b"},{"src":"c","dst":"b"}]`); !errors.Is(err, ErrInvalidColumnMapping) {
		t.Fatalf("dst 重复应拒绝, got %v", err)
	}
}

// TestResolveColumns 目标列解析：直通 / 投影 / 错误
func TestResolveColumns(t *testing.T) {
	cols := []string{"id", "name", "amount"}
	types := []string{"BIGINT", "VARCHAR", "DECIMAL"}

	// 缺省直通：列名直通、类型按数据库类型名推断
	got, idx, err := ResolveColumns(cols, types, nil)
	if err != nil {
		t.Fatalf("直通解析不应报错: %v", err)
	}
	want := []TargetColumn{
		{Name: "id", Type: "bigint"},
		{Name: "name", Type: "varchar(255)"},
		{Name: "amount", Type: "decimal(18,4)"},
	}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("直通列不符: %+v", got)
	}
	if len(idx) != 3 || idx[0] != 0 || idx[1] != 1 || idx[2] != 2 {
		t.Fatalf("直通下标不符: %v", idx)
	}

	// 投影：调序+改名+缺省类型推断
	mapping := []ColumnMapping{
		{Src: "amount", Dst: "amt", Type: "decimal"},
		{Src: "id", Dst: "oid"}, // type 缺省 → 按 BIGINT 推断
	}
	got, idx, err = ResolveColumns(cols, types, mapping)
	if err != nil {
		t.Fatalf("投影解析不应报错: %v", err)
	}
	if got[0].Name != "amt" || got[0].Type != "decimal(18,4)" || got[1].Name != "oid" || got[1].Type != "bigint" {
		t.Fatalf("投影列不符: %+v", got)
	}
	if idx[0] != 2 || idx[1] != 0 {
		t.Fatalf("投影下标不符: %v", idx)
	}

	// 映射列在结果集中不存在
	if _, _, err := ResolveColumns(cols, types, []ColumnMapping{{Src: "nope", Dst: "x"}}); err == nil {
		t.Fatalf("缺失映射列应报错")
	}
	// 表达式列未起别名
	if _, _, err := ResolveColumns([]string{"count(*)"}, []string{"BIGINT"}, nil); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("非法列名应拒绝, got %v", err)
	}
	// 空结果集
	if _, _, err := ResolveColumns(nil, nil, nil); err == nil {
		t.Fatalf("空结果集应报错")
	}
}

// TestInferSqlType 数据库类型名 → 物理列类型推断
func TestInferSqlType(t *testing.T) {
	cases := map[string]string{
		"BIGINT":     "bigint",
		"int":        "bigint",
		"TINYINT":    "bigint", // 不特判 tinyint(1)，tinyint(1) 布尔由 mapping.type=bool 显式指定
		"DECIMAL":    "decimal(18,4)",
		"NEWDECIMAL": "decimal(18,4)",
		"NUMERIC":    "decimal(18,4)",
		"BOOL":       "tinyint(1)",
		"BOOLEAN":    "tinyint(1)",
		"DATETIME":   "datetime",
		"TIMESTAMP":  "datetime",
		"DATE":       "datetime",
		"VARCHAR":    "varchar(255)",
		"CHAR":       "varchar(255)",
		"ENUM":       "varchar(255)",
		"SET":        "varchar(255)",
		"TEXT":       "text",
		"LONGTEXT":   "text",
		"BLOB":       "text",
		"JSON":       "text",
		"":           "varchar(255)", // 未知/获取失败回落
		"WEIRD":      "varchar(255)",
	}
	for in, want := range cases {
		if got := InferSqlType(in); got != want {
			t.Fatalf("InferSqlType(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestNormalizeBatchSize 分批大小归一
func TestNormalizeBatchSize(t *testing.T) {
	cases := map[int64]int64{
		0:     model.DefaultSyncBatchSize,
		-5:    model.DefaultSyncBatchSize,
		500:   500,
		1000:  1000,
		10000: model.MaxSyncBatchSize,
		10001: model.MaxSyncBatchSize,
	}
	for in, want := range cases {
		if got := model.NormalizeBatchSize(in); got != want {
			t.Fatalf("NormalizeBatchSize(%d) = %d, want %d", in, got, want)
		}
	}
}

// ============ 分批主循环 mock ============

// fakeSource 源库 mock：预置每批行数据（读空后持续返回空批次），可指定第几批读取失败
type fakeSource struct {
	cols    []string
	types   []string
	batches [][][]any // 每批返回的行
	errAt   int       // 第几批读取失败（1 起，0=不失败）
	calls   int
}

func (f *fakeSource) ReadBatch(_ context.Context, _ string, offset, limit int64) (*BatchData, error) {
	f.calls++
	batchIdx := int(offset / limit)
	if f.errAt == batchIdx+1 {
		return nil, errors.New("read boom")
	}
	if batchIdx < len(f.batches) {
		return &BatchData{Columns: f.cols, Types: f.types, Rows: f.batches[batchIdx]}, nil
	}
	return &BatchData{Columns: f.cols, Types: f.types}, nil
}

// fakeTarget 目标库 mock：记录建表与写入，可指定第几批写入失败
type fakeTarget struct {
	ensureCalled int
	lastCols     []TargetColumn
	rows         [][]any
	errAtBatch   int // 第几批写入失败（1 起，0=不失败）
	batch        int
}

func (t *fakeTarget) EnsureTable(_ context.Context, _ string, cols []TargetColumn) error {
	t.ensureCalled++
	t.lastCols = cols
	return nil
}

func (t *fakeTarget) WriteBatch(_ context.Context, _ string, _ []TargetColumn, r [][]any) (int64, error) {
	t.batch++
	if t.errAtBatch == t.batch {
		return 0, errors.New("write boom")
	}
	t.rows = append(t.rows, r...)
	return int64(len(r)), nil
}

// TestRunSyncLoopSuccess 全量跑完：批数/读写累计/建表一次
func TestRunSyncLoopSuccess(t *testing.T) {
	src := &fakeSource{
		cols:  []string{"id", "name"},
		types: []string{"BIGINT", "VARCHAR"},
		batches: [][][]any{
			{{int64(1), []byte("a")}, {int64(2), []byte("b")}},
			{{int64(3), []byte("c")}, {int64(4), []byte("d")}},
			{{int64(5), []byte("e")}}, // 不足一批 → 收尾
		},
	}
	tgt := &fakeTarget{}
	stats, err := RunSyncLoop(context.Background(), src, tgt, SyncOptions{
		SourceSql: "select id,name from t", QualifiedTable: "`dwd_t`", BatchSize: 2,
	})
	if err != nil {
		t.Fatalf("分批同步不应报错: %v", err)
	}
	if stats.ReadRows != 5 || stats.WriteRows != 5 || stats.Batches != 3 {
		t.Fatalf("累计不符: %+v", stats)
	}
	if tgt.ensureCalled != 1 {
		t.Fatalf("建表应恰好执行一次, got %d", tgt.ensureCalled)
	}
	if len(tgt.lastCols) != 2 || tgt.lastCols[0].Name != "id" || tgt.lastCols[0].Type != "bigint" {
		t.Fatalf("建表列不符: %+v", tgt.lastCols)
	}
	if len(tgt.rows) != 5 || string(tgt.rows[4][1].([]byte)) != "e" {
		t.Fatalf("写入行不符: %+v", tgt.rows)
	}
}

// TestRunSyncLoopEmptySource 源无数据：直接成功收尾，不建表
func TestRunSyncLoopEmptySource(t *testing.T) {
	src := &fakeSource{cols: []string{"id"}, types: []string{"BIGINT"}}
	tgt := &fakeTarget{}
	stats, err := RunSyncLoop(context.Background(), src, tgt,
		SyncOptions{SourceSql: "select id from t", QualifiedTable: "`t`"})
	if err != nil {
		t.Fatalf("空源不应报错: %v", err)
	}
	if stats.ReadRows != 0 || stats.WriteRows != 0 || stats.Batches != 0 {
		t.Fatalf("空源累计应为 0: %+v", stats)
	}
	if tgt.ensureCalled != 0 {
		t.Fatalf("空源不应建表, got %d", tgt.ensureCalled)
	}
}

// TestRunSyncLoopWriteFail 写目标第 2 批失败：已提交批保留（读 2 写 2 批数 1），
// 失败批号 = Batches+1 = 2
func TestRunSyncLoopWriteFail(t *testing.T) {
	src := &fakeSource{
		cols:  []string{"id"},
		types: []string{"BIGINT"},
		batches: [][][]any{
			{{int64(1)}, {int64(2)}},
			{{int64(3)}, {int64(4)}},
			{{int64(5)}, {int64(6)}},
		},
	}
	tgt := &fakeTarget{errAtBatch: 2}
	stats, err := RunSyncLoop(context.Background(), src, tgt,
		SyncOptions{SourceSql: "select id from t", QualifiedTable: "`t`", BatchSize: 2})
	if err == nil || !strings.Contains(err.Error(), "写入目标表失败") {
		t.Fatalf("应返回写入失败错误, got %v", err)
	}
	if stats.Batches != 1 || stats.ReadRows != 2 || stats.WriteRows != 2 {
		t.Fatalf("已提交批应保留: %+v", stats)
	}
	if stats.Batches+1 != 2 {
		t.Fatalf("失败批号应为 2, got %d", stats.Batches+1)
	}
	if len(tgt.rows) != 2 {
		t.Fatalf("失败批数据不应计入已写入: %+v", tgt.rows)
	}
}

// TestRunSyncLoopReadFail 读源第 1 批失败：零累计
func TestRunSyncLoopReadFail(t *testing.T) {
	src := &fakeSource{cols: []string{"id"}, types: []string{"BIGINT"}, errAt: 1}
	tgt := &fakeTarget{}
	stats, err := RunSyncLoop(context.Background(), src, tgt,
		SyncOptions{SourceSql: "select id from t", QualifiedTable: "`t`"})
	if err == nil || !strings.Contains(err.Error(), "读取源数据失败") {
		t.Fatalf("应返回读取失败错误, got %v", err)
	}
	if stats.ReadRows != 0 || stats.WriteRows != 0 || stats.Batches != 0 {
		t.Fatalf("零累计: %+v", stats)
	}
}

// TestRunSyncLoopMapping 投影模式：行数据按映射取值
func TestRunSyncLoopMapping(t *testing.T) {
	src := &fakeSource{
		cols:  []string{"id", "name", "user_id"},
		types: []string{"BIGINT", "VARCHAR", "BIGINT"},
		batches: [][][]any{
			{{int64(1), []byte("a"), int64(11)}},
		},
	}
	tgt := &fakeTarget{}
	mapping := []ColumnMapping{
		{Src: "user_id", Dst: "uid", Type: "int"},
		{Src: "id", Dst: "oid"},
	}
	stats, err := RunSyncLoop(context.Background(), src, tgt, SyncOptions{
		SourceSql: "select id,name,user_id from t", QualifiedTable: "`t`",
		Mapping: mapping, BatchSize: 10,
	})
	if err != nil {
		t.Fatalf("投影同步不应报错: %v", err)
	}
	if stats.WriteRows != 1 {
		t.Fatalf("写入行数不符: %+v", stats)
	}
	if len(tgt.lastCols) != 2 || tgt.lastCols[0].Name != "uid" || tgt.lastCols[1].Name != "oid" {
		t.Fatalf("投影列不符: %+v", tgt.lastCols)
	}
	if len(tgt.rows) != 1 || tgt.rows[0][0] != int64(11) || tgt.rows[0][1] != int64(1) {
		t.Fatalf("投影取值不符: %+v", tgt.rows)
	}
}

// TestRunSyncLoopExactBatches 整批对齐：4 行/批 2 → 2 批后再读一次空批收尾
func TestRunSyncLoopExactBatches(t *testing.T) {
	src := &fakeSource{
		cols:  []string{"id"},
		types: []string{"BIGINT"},
		batches: [][][]any{
			{{int64(1)}, {int64(2)}},
			{{int64(3)}, {int64(4)}},
		},
	}
	tgt := &fakeTarget{}
	stats, err := RunSyncLoop(context.Background(), src, tgt,
		SyncOptions{SourceSql: "select id from t", QualifiedTable: "`t`", BatchSize: 2})
	if err != nil {
		t.Fatalf("整批同步不应报错: %v", err)
	}
	if stats.Batches != 2 || stats.ReadRows != 4 || src.calls != 3 {
		t.Fatalf("整批累计不符: %+v calls=%d", stats, src.calls)
	}
}

// TestNextTaskStatus 内置同步后的任务状态流转
func TestNextTaskStatus(t *testing.T) {
	cases := []struct {
		local     string
		runStatus string
		want      string
	}{
		{model.TaskStatusDraft, model.SyncLogStatusSuccess, model.TaskStatusDraft},
		{model.TaskStatusDeployed, model.SyncLogStatusSuccess, model.TaskStatusDeployed},
		{model.TaskStatusFailed, model.SyncLogStatusSuccess, model.TaskStatusDeployed},
		{model.TaskStatusRunning, model.SyncLogStatusSuccess, model.TaskStatusDeployed},
		{model.TaskStatusDraft, model.SyncLogStatusFailed, model.TaskStatusFailed},
		{model.TaskStatusDeployed, model.SyncLogStatusFailed, model.TaskStatusFailed},
	}
	for _, c := range cases {
		if got := NextTaskStatus(c.local, c.runStatus); got != c.want {
			t.Fatalf("NextTaskStatus(%q,%q) = %q, want %q", c.local, c.runStatus, got, c.want)
		}
	}
}

// TestQualifyTable 全限定名拼装与白名单
func TestQualifyTable(t *testing.T) {
	if got, err := QualifyTable("", "t1"); err != nil || got != "`t1`" {
		t.Fatalf("QualifyTable(\"\","+"t1\") = %q,%v", got, err)
	}
	if got, err := QualifyTable("db", "t1"); err != nil || got != "`db`.`t1`" {
		t.Fatalf("QualifyTable(db,t1) = %q,%v", got, err)
	}
	if _, err := QualifyTable("db; drop", "t1"); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("库名注入应拒绝, got %v", err)
	}
	if _, err := QualifyTable("db", "t drop"); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("表名注入应拒绝, got %v", err)
	}
}

// TestBuildCreateTargetDDL 自动建表 DDL 拼装
func TestBuildCreateTargetDDL(t *testing.T) {
	ddl := buildCreateTargetDDL("`db`.`t1`", []TargetColumn{
		{Name: "id", Type: "bigint"},
		{Name: "name", Type: "varchar(255)"},
	})
	for _, part := range []string{
		"create table if not exists `db`.`t1`",
		"`id` bigint unsigned NOT NULL AUTO_INCREMENT",
		"`id` bigint DEFAULT NULL",
		"`name` varchar(255) DEFAULT NULL",
		"PRIMARY KEY (`id`)",
		"ENGINE=InnoDB",
	} {
		if !strings.Contains(ddl, part) {
			t.Fatalf("DDL 缺少 %q: %s", part, ddl)
		}
	}
}

// TestBuildReplaceSQL replace into 语句与占位符数量
func TestBuildReplaceSQL(t *testing.T) {
	query, n := buildReplaceSQL("`t1`", []TargetColumn{{Name: "oid"}, {Name: "uname"}}, 3)
	if !strings.Contains(query, "replace into `t1` (`oid`,`uname`) values") {
		t.Fatalf("replace 语句不符: %s", query)
	}
	if n != 6 { // 2 列 × 3 行
		t.Fatalf("占位符总数应为 6, got %d", n)
	}
	if got := strings.Count(query, "(?,?)"); got != 3 {
		t.Fatalf("values 行数应为 3, got %d: %s", got, query)
	}
}

// TestProjectRows 行投影：有下标按位取值，无下标原样直通
func TestProjectRows(t *testing.T) {
	rows := [][]any{{int64(1), "a", int64(2)}, {int64(3), "b", int64(4)}}
	got := projectRows(rows, nil)
	if len(got) != 2 || len(got[0]) != 3 {
		t.Fatalf("直通应原样返回: %+v", got)
	}
	got = projectRows(rows, []int{2, 0})
	if got[0][0] != int64(2) || got[0][1] != int64(1) || got[1][0] != int64(4) {
		t.Fatalf("投影取值不符: %+v", got)
	}
}

// TestTruncateRunes 按字符截断（中文按 1 字符计）
func TestTruncateRunes(t *testing.T) {
	if got := truncateRunes("你好世界", 3); got != "你好世" {
		t.Fatalf("truncateRunes = %q", got)
	}
	if got := truncateRunes("abc", 10); got != "abc" {
		t.Fatalf("truncateRunes = %q", got)
	}
	if got := truncateRunes(strings.Repeat("长", syncMsgMaxLen+10), syncMsgMaxLen); len([]rune(got)) != syncMsgMaxLen {
		t.Fatalf("超长 msg 应截断到 %d 字符, got %d", syncMsgMaxLen, len([]rune(got)))
	}
}
