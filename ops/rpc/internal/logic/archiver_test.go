package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// mockArchiveDB ArchiveDB 的单测 mock：按预置脚本逐批返回结果，可注入指定批失败
type mockArchiveDB struct {
	batches    []BatchResult // 每批返回（耗尽后返回 0 行）
	errAt      int           // 第 errAt 批（1 起算）返回错误；0 表示不失败
	alwaysFull bool          // 恒返回满批（模拟永远取不完数据，验证 MaxBatches 上限）
	gotSpecs   []BatchSpec   // 记录每批收到的执行参数
}

func (m *mockArchiveDB) ArchiveBatch(_ context.Context, spec BatchSpec) (BatchResult, error) {
	m.gotSpecs = append(m.gotSpecs, spec)
	if m.errAt > 0 && len(m.gotSpecs) == m.errAt {
		return BatchResult{}, fmt.Errorf("mock 第 %d 批执行失败", len(m.gotSpecs))
	}
	if m.alwaysFull {
		res := BatchResult{Archived: spec.Limit}
		if spec.DeleteSource {
			res.Deleted = spec.Limit
		}
		return res, nil
	}
	if len(m.gotSpecs) <= len(m.batches) {
		return m.batches[len(m.gotSpecs)-1], nil
	}
	return BatchResult{}, nil
}

// 测试 delete 模式主流程：逐批 insert+delete 推进，0 行自然收尾，累计正确
func TestRunArchiveLoopDeleteMode(t *testing.T) {
	m := &mockArchiveDB{batches: []BatchResult{
		{Archived: 1000, Deleted: 1000},
		{Archived: 1000, Deleted: 1000},
		{Archived: 500, Deleted: 500},
		{Archived: 0},
	}}
	stats, err := RunArchiveLoop(context.Background(), m, ArchiveOptions{
		SourceTable:  "mall.orders",
		ArchiveTable: "mall.orders_arch",
		Condition:    "where created_at < '2024-01-01'",
		BatchSize:    1000,
		DeleteSource: true,
	})
	if err != nil {
		t.Fatalf("RunArchiveLoop error: %v", err)
	}
	if stats.Archived != 2500 || stats.Deleted != 2500 {
		t.Fatalf("累计 = archived %d / deleted %d, want 2500/2500", stats.Archived, stats.Deleted)
	}
	if stats.Batches != 4 {
		t.Fatalf("批数 = %d, want 4", stats.Batches)
	}
	// delete 模式恒从头部取：不带 offset、每批都带删源标记与批大小
	for i, spec := range m.gotSpecs {
		if spec.Offset != -1 {
			t.Fatalf("delete 模式第 %d 批不应带 offset, got %d", i+1, spec.Offset)
		}
		if !spec.DeleteSource {
			t.Fatalf("delete 模式第 %d 批 DeleteSource 应为 true", i+1)
		}
		if spec.Limit != 1000 {
			t.Fatalf("delete 模式第 %d 批 Limit = %d, want 1000", i+1, spec.Limit)
		}
		if spec.SourceTable != "mall.orders" || spec.ArchiveTable != "mall.orders_arch" {
			t.Fatalf("第 %d 批表名错误: %v", i+1, spec)
		}
	}
}

// 测试纯复制模式（仅插入）：offset 逐批翻页，避免每批反复归档同一批行
func TestRunArchiveLoopCopyModeOffset(t *testing.T) {
	m := &mockArchiveDB{batches: []BatchResult{
		{Archived: 1000},
		{Archived: 1000},
		{Archived: 300},
		{Archived: 0},
	}}
	stats, err := RunArchiveLoop(context.Background(), m, ArchiveOptions{
		SourceTable:  "orders",
		ArchiveTable: "orders_arch",
		Condition:    "where status = 1",
		BatchSize:    1000,
		DeleteSource: false,
	})
	if err != nil {
		t.Fatalf("RunArchiveLoop error: %v", err)
	}
	if stats.Archived != 2300 || stats.Deleted != 0 {
		t.Fatalf("累计 = archived %d / deleted %d, want 2300/0", stats.Archived, stats.Deleted)
	}
	// offset 应按 0 → 1000 → 2000 → 2300 递进
	wantOffsets := []int64{0, 1000, 2000, 2300}
	for i, spec := range m.gotSpecs {
		if spec.Offset != wantOffsets[i] {
			t.Fatalf("第 %d 批 offset = %d, want %d", i+1, spec.Offset, wantOffsets[i])
		}
		if spec.DeleteSource {
			t.Fatalf("纯复制模式第 %d 批 DeleteSource 应为 false", i+1)
		}
	}
}

// 测试失败中断：某批失败立即停止，失败前已提交批的累计值保留，错误上抛
func TestRunArchiveLoopStopsOnBatchFailure(t *testing.T) {
	m := &mockArchiveDB{
		batches: []BatchResult{{Archived: 1000, Deleted: 1000}},
		errAt:   2, // 第 2 批失败
	}
	stats, err := RunArchiveLoop(context.Background(), m, ArchiveOptions{
		SourceTable: "orders", ArchiveTable: "orders_arch",
		Condition: "where status = 1", BatchSize: 1000, DeleteSource: true,
	})
	if err == nil {
		t.Fatalf("第 2 批失败应上抛错误")
	}
	// 已提交的第 1 批保留在累计值里
	if stats.Archived != 1000 || stats.Deleted != 1000 || stats.Batches != 1 {
		t.Fatalf("失败中断后累计 = %+v, want archived 1000/deleted 1000/batches 1", stats)
	}
	if len(m.gotSpecs) != 2 {
		t.Fatalf("应尝试执行 2 批, got %d", len(m.gotSpecs))
	}

	// 首批即失败：累计为 0，错误上抛
	m2 := &mockArchiveDB{errAt: 1}
	stats2, err2 := RunArchiveLoop(context.Background(), m2, ArchiveOptions{
		SourceTable: "orders", ArchiveTable: "orders_arch", BatchSize: 1000, DeleteSource: true,
	})
	if err2 == nil || stats2.Archived != 0 || stats2.Batches != 0 {
		t.Fatalf("首批失败应返回零累计与错误, got %+v, %v", stats2, err2)
	}
}

// 测试首批 0 行：无可归档数据，只执行一批即收尾
func TestRunArchiveLoopEmptySource(t *testing.T) {
	m := &mockArchiveDB{batches: []BatchResult{{Archived: 0}}}
	stats, err := RunArchiveLoop(context.Background(), m, ArchiveOptions{
		SourceTable: "orders", ArchiveTable: "orders_arch",
		BatchSize: 1000, DeleteSource: true,
	})
	if err != nil {
		t.Fatalf("RunArchiveLoop error: %v", err)
	}
	if stats.Archived != 0 || stats.Deleted != 0 || stats.Batches != 1 {
		t.Fatalf("空源累计 = %+v, want 零值且批数为 1", stats)
	}
}

// 测试 MaxBatches 安全上限：恒取得到数据时按上限截停，不进入死循环
func TestRunArchiveLoopMaxBatchesCap(t *testing.T) {
	m := &mockArchiveDB{alwaysFull: true}
	stats, err := RunArchiveLoop(context.Background(), m, ArchiveOptions{
		SourceTable: "orders", ArchiveTable: "orders_arch",
		BatchSize: 1000, DeleteSource: true, MaxBatches: 5,
	})
	if err != nil {
		t.Fatalf("按上限截停不应报错, got %v", err)
	}
	if stats.Batches != 5 || stats.Archived != 5000 || stats.Deleted != 5000 {
		t.Fatalf("上限截停累计 = %+v, want batches 5/archived 5000/deleted 5000", stats)
	}
}

// 测试缺省参数兜底：batch_size<=0 取 DefaultBatchSize
func TestRunArchiveLoopDefaults(t *testing.T) {
	m := &mockArchiveDB{batches: []BatchResult{{Archived: 0}}}
	if _, err := RunArchiveLoop(context.Background(), m, ArchiveOptions{
		SourceTable: "orders", ArchiveTable: "orders_arch", BatchSize: 0,
	}); err != nil {
		t.Fatalf("RunArchiveLoop error: %v", err)
	}
	if m.gotSpecs[0].Limit != DefaultBatchSize {
		t.Fatalf("batch_size 缺省应取 %d, got %d", DefaultBatchSize, m.gotSpecs[0].Limit)
	}
}

// 测试单批 SQL 生成：insert select / delete 与条件、offset 的拼接
func TestBuildArchiveSQL(t *testing.T) {
	// 常规 delete 模式插入（不带 offset）
	insSQL, insArgs := buildInsertSQL(BatchSpec{
		SourceTable: "mall.orders", ArchiveTable: "mall.orders_arch",
		Condition: "where created_at < '2024-01-01'", Offset: -1, Limit: 1000,
	})
	wantIns := "insert into mall.orders_arch select * from mall.orders where created_at < '2024-01-01' limit ?"
	if insSQL != wantIns {
		t.Fatalf("insert sql = %q, want %q", insSQL, wantIns)
	}
	if len(insArgs) != 1 || insArgs[0] != int64(1000) {
		t.Fatalf("insert args = %v, want [1000]", insArgs)
	}

	// 纯复制模式带 offset 翻页
	insSQL, insArgs = buildInsertSQL(BatchSpec{
		SourceTable: "orders", ArchiveTable: "orders_arch",
		Condition: "where status = 1", Offset: 500, Limit: 1000,
	})
	wantIns = "insert into orders_arch select * from orders where status = 1 limit ?, ?"
	if insSQL != wantIns {
		t.Fatalf("insert(offset) sql = %q, want %q", insSQL, wantIns)
	}
	if len(insArgs) != 2 || insArgs[0] != int64(500) || insArgs[1] != int64(1000) {
		t.Fatalf("insert(offset) args = %v, want [500 1000]", insArgs)
	}

	// 空条件（整表归档）
	insSQL, _ = buildInsertSQL(BatchSpec{
		SourceTable: "orders", ArchiveTable: "orders_arch", Offset: -1, Limit: 1000,
	})
	if insSQL != "insert into orders_arch select * from orders limit ?" {
		t.Fatalf("空条件 insert sql = %q", insSQL)
	}

	// 删源 SQL
	delSQL, delArgs := buildDeleteSQL(BatchSpec{
		SourceTable: "mall.orders", Condition: "where created_at < '2024-01-01'", Limit: 1000,
	})
	wantDel := "delete from mall.orders where created_at < '2024-01-01' limit ?"
	if delSQL != wantDel {
		t.Fatalf("delete sql = %q, want %q", delSQL, wantDel)
	}
	if len(delArgs) != 1 || delArgs[0] != int64(1000) {
		t.Fatalf("delete args = %v, want [1000]", delArgs)
	}
	// 空条件删源
	delSQL, _ = buildDeleteSQL(BatchSpec{SourceTable: "orders", Limit: 100})
	if delSQL != "delete from orders limit ?" {
		t.Fatalf("空条件 delete sql = %q", delSQL)
	}
}

// 测试归档条件校验：合法 where 片段（含任务书指定样例）全部放行
func TestValidateConditionSQLPass(t *testing.T) {
	valid := []string{
		"",                                // 空 = 整表归档
		"   ",                             // 纯空白视同空
		"where created_at < '2024-01-01'", // 任务书指定样例
		"WHERE status = 1 AND created_at >= '2023-01-01 00:00:00'", // 大小写不敏感
		"where remark = 'a;b'",                         // 分号在字面量内不构成注入
		"where note = 'it''s -- ok # fine /* x */'",    // 注释符在字面量内
		"where user_id in (select id from black_list)", // 子查询放行
		"where created = 1",                            // 词边界：created 不误伤 create
		"where name like '%drop%'",                     // 关键字在字面量内
	}
	for _, cond := range valid {
		if err := ValidateConditionSQL(cond); err != nil {
			t.Fatalf("条件 %q 应通过校验, got %v", cond, err)
		}
	}
}

// 测试归档条件黑名单：分号/注释/drop 等关键字一律拒绝
func TestValidateConditionSQLReject(t *testing.T) {
	rejects := []struct {
		cond string
		why  string
	}{
		{"created_at < '2024-01-01'", "缺 where 前缀"},
		{"select 1", "非 where 片段"},
		{"where 1=1; drop table t", "分号 + drop"},
		{"where 1=1;truncate table t;", "分号 + truncate"},
		{"where a = 1 -- 备注", "行注释 --"},
		{"where a = 1--x", "行注释 -- 紧贴"},
		{"where a=1 # 注释", "行注释 #"},
		{"where a=1 /* 块注释 */", "块注释"},
		{"where 1=1 union all select 1", "union 注入"},
		{"where sleep(3)", "sleep 注入"},
		{"where a=1 into outfile '/tmp/x'", "into outfile"},
		{"where load_file('/etc/passwd')", "load_file"},
		{"where 1=1 or 1=1; delete from t", "多语句 delete"},
		{"where a = 'unterminated", "引号未闭合"},
		{"where a=1 and b=2; --", "尾部分号 + 注释"},
		{"where 1=1 and drop", "裸 drop 关键字"},
		{"where created_at < '2024-01-01' and 1=1; alter table t", "alter 追加"},
	}
	for _, c := range rejects {
		if err := ValidateConditionSQL(c.cond); err == nil {
			t.Fatalf("条件 %q（%s）应被拒绝", c.cond, c.why)
		}
	}
	// 错误语义：前缀问题 vs 黑名单问题可区分
	if err := ValidateConditionSQL("select 1"); !errors.Is(err, ErrInvalidCondition) {
		t.Fatalf("缺 where 前缀应报 ErrInvalidCondition, got %v", err)
	}
	if err := ValidateConditionSQL("where 1=1; drop table t"); !errors.Is(err, ErrInvalidCondition) {
		t.Fatalf("分号应报 ErrInvalidCondition, got %v", err)
	}
	// 超长拒绝（varchar(1024)）
	long := "where col = '" + strings.Repeat("a", 1100) + "'"
	if err := ValidateConditionSQL(long); !errors.Is(err, ErrConditionTooLong) {
		t.Fatalf("超长条件应报 ErrConditionTooLong, got %v", err)
	}
}

// 测试标识符白名单正则的边界
func TestValidIdentifier(t *testing.T) {
	valid := []string{"orders", "t_1", "_tmp", "OrderItems", "orders_arch", "a"}
	for _, s := range valid {
		if !ValidIdentifier(s) {
			t.Fatalf("%q 应为合法标识符", s)
		}
	}
	invalid := []string{
		"", "orders; drop table x", "a b", "a-b", "a.b", "`a`", "'a'",
		"a*", "中文名", "orders--", "orders\n",
	}
	for _, s := range invalid {
		if ValidIdentifier(s) {
			t.Fatalf("%q 应为非法标识符", s)
		}
	}
}

// 测试库前缀生成
func TestQualifyTable(t *testing.T) {
	got, err := QualifyTable("mall", "orders")
	if err != nil || got != "mall.orders" {
		t.Fatalf("QualifyTable = %q, %v", got, err)
	}
	got, err = QualifyTable("", "orders")
	if err != nil || got != "orders" {
		t.Fatalf("空库名应返回裸表名, got %q, %v", got, err)
	}
	if _, err := QualifyTable("mall", "orders; drop table x"); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("表名注入应被拒绝, got %v", err)
	}
	if _, err := QualifyTable("mall.orders", "orders"); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("库名含点号应被拒绝（前缀由 db 参数承担）, got %v", err)
	}
}
