package logic

import (
	"encoding/json"
	"reflect"
	"testing"

	"datacenter/metadata/model"
)

// 测试快照无差异：字段全相同时返回空数组
func TestDiffEntitySnapshotNoDiff(t *testing.T) {
	before := &entitySnapshot{Name: "t1", Comment: "c", Engine: "InnoDB", RowsCnt: 10,
		PropsJson: "{}", Classification: "L2"}
	after := &entitySnapshot{Name: "t1", Comment: "c", Engine: "InnoDB", RowsCnt: 10,
		PropsJson: "{}", Classification: "L2"}

	diffs := diffEntitySnapshot(before, after)
	if len(diffs) != 0 {
		t.Fatalf("diffs = %v, want empty", diffs)
	}
}

// 测试全字段差异：六个对比字段全部变化时按固定顺序输出
func TestDiffEntitySnapshotAllDiff(t *testing.T) {
	before := &entitySnapshot{Name: "t1", Comment: "旧注释", Engine: "MyISAM", RowsCnt: 10,
		PropsJson: `{"a":1}`, Classification: "L2"}
	after := &entitySnapshot{Name: "t2", Comment: "新注释", Engine: "InnoDB", RowsCnt: 20,
		PropsJson: `{"a":2}`, Classification: "L3"}

	diffs := diffEntitySnapshot(before, after)
	want := []FieldDiff{
		{Field: "name", Before: "t1", After: "t2"},
		{Field: "comment", Before: "旧注释", After: "新注释"},
		{Field: "engine", Before: "MyISAM", After: "InnoDB"},
		{Field: "rows_cnt", Before: "10", After: "20"},
		{Field: "props_json", Before: `{"a":1}`, After: `{"a":2}`},
		{Field: "classification", Before: "L2", After: "L3"},
	}
	if !reflect.DeepEqual(diffs, want) {
		t.Fatalf("diffs = %+v, want %+v", diffs, want)
	}
}

// 测试部分字段差异：只输出变化的字段
func TestDiffEntitySnapshotPartialDiff(t *testing.T) {
	before := &entitySnapshot{Name: "t1", Comment: "c", Engine: "InnoDB", RowsCnt: 10,
		PropsJson: "{}", Classification: "L2"}
	after := &entitySnapshot{Name: "t1", Comment: "c", Engine: "InnoDB", RowsCnt: 99,
		PropsJson: "{}", Classification: "L4"}

	diffs := diffEntitySnapshot(before, after)
	want := []FieldDiff{
		{Field: "rows_cnt", Before: "10", After: "99"},
		{Field: "classification", Before: "L2", After: "L4"},
	}
	if !reflect.DeepEqual(diffs, want) {
		t.Fatalf("diffs = %+v, want %+v", diffs, want)
	}
}

// 测试实体 -> 快照 -> JSON -> 快照 的往返一致性
func TestEntitySnapshotRoundTrip(t *testing.T) {
	entity := &model.Entity{
		Id: 1, EntityType: "table", Fqn: "mysql://db.t1", Name: "t1",
		Comment: "订单表", Engine: "InnoDB", RowsCnt: 100,
		PropsJson: `{"charset":"utf8mb4"}`, Classification: "L3",
		Owner: "jack", DsId: 7, Beid: 1,
	}
	snap := newEntitySnapshot(entity)
	if snap.EntityType != "table" || snap.Fqn != "mysql://db.t1" || snap.Owner != "jack" {
		t.Fatalf("snapshot = %+v", snap)
	}

	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	parsed, err := parseEntitySnapshot(string(data))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if !reflect.DeepEqual(snap, parsed) {
		t.Fatalf("parsed = %+v, want %+v", parsed, snap)
	}
}

// 测试非法快照 JSON 返回错误
func TestParseEntitySnapshotInvalidJson(t *testing.T) {
	if _, err := parseEntitySnapshot("not-json"); err == nil {
		t.Fatalf("want error for invalid json, got nil")
	}
}
