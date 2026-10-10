package logic

import (
	"testing"
)

// 期望类型串生成：varchar(50)/decimal(10,2)/基础类型名
func TestExpectType(t *testing.T) {
	cases := []struct {
		name     string
		dataType string
		length   int32
		prec     int32
		want     string
	}{
		{"varchar 带长度", DataTypeVarchar, 50, 0, "varchar(50)"},
		{"decimal 总精度+小数位", DataTypeDecimal, 10, 2, "decimal(10,2)"},
		{"int 无长度修饰", DataTypeInt, 0, 0, "int"},
		{"bigint", DataTypeBigint, 0, 0, "bigint"},
		{"date", DataTypeDate, 0, 0, "date"},
		{"datetime", DataTypeDatetime, 0, 0, "datetime"},
		{"text", DataTypeText, 0, 0, "text"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ExpectType(c.dataType, c.length, c.prec); got != c.want {
				t.Fatalf("ExpectType(%s,%d,%d) = %q, want %q", c.dataType, c.length, c.prec, got, c.want)
			}
		})
	}
}

// 类型匹配纯函数：varchar 长度严格一致 / decimal 总精度与小数位 / 基础类型等价
func TestTypeMatched(t *testing.T) {
	cases := []struct {
		name   string
		std    StdColumn
		actual ActualColumn
		want   bool
	}{
		{
			// 标准要求 varchar(50)，实际 varchar(50) → 匹配
			name:   "varchar(50) 标准vs实际一致",
			std:    StdColumn{Name: "user_name", DataType: DataTypeVarchar, Length: 50},
			actual: ActualColumn{Name: "user_name", DataType: "varchar", ColumnType: "varchar(50)", CharLen: 50},
			want:   true,
		},
		{
			// varchar(50) vs varchar(100) → 长度不一致不匹配
			name:   "varchar 长度不一致",
			std:    StdColumn{Name: "user_name", DataType: DataTypeVarchar, Length: 50},
			actual: ActualColumn{Name: "user_name", DataType: "varchar", ColumnType: "varchar(100)", CharLen: 100},
			want:   false,
		},
		{
			// varchar vs text → 基础类型不同不匹配
			name:   "varchar 实际为 text",
			std:    StdColumn{Name: "user_name", DataType: DataTypeVarchar, Length: 50},
			actual: ActualColumn{Name: "user_name", DataType: "text", ColumnType: "text"},
			want:   false,
		},
		{
			// decimal(10,2)：总精度 10、小数位 2 全部一致 → 匹配
			name:   "decimal(10,2) 标准vs实际一致",
			std:    StdColumn{Name: "amount", DataType: DataTypeDecimal, Length: 10, Precision: 2},
			actual: ActualColumn{Name: "amount", DataType: "decimal", ColumnType: "decimal(10,2)", NumPrecision: 10, NumScale: 2},
			want:   true,
		},
		{
			// decimal(10,2) vs decimal(12,2)：总精度不一致不匹配
			name:   "decimal 总精度不一致",
			std:    StdColumn{Name: "amount", DataType: DataTypeDecimal, Length: 10, Precision: 2},
			actual: ActualColumn{Name: "amount", DataType: "decimal", ColumnType: "decimal(12,2)", NumPrecision: 12, NumScale: 2},
			want:   false,
		},
		{
			// decimal(10,2) vs decimal(10,0)：小数位不一致不匹配
			name:   "decimal 小数位不一致",
			std:    StdColumn{Name: "amount", DataType: DataTypeDecimal, Length: 10, Precision: 2},
			actual: ActualColumn{Name: "amount", DataType: "decimal", ColumnType: "decimal(10,0)", NumPrecision: 10},
			want:   false,
		},
		{
			// int 基础类型一致即匹配（实际可带 unsigned 修饰，DATA_TYPE 仍为 int）
			name:   "int 匹配",
			std:    StdColumn{Name: "age", DataType: DataTypeInt},
			actual: ActualColumn{Name: "age", DataType: "int", ColumnType: "int unsigned"},
			want:   true,
		},
		{
			// int vs bigint → 不匹配（标准侧严格等价，不做可容纳性放宽）
			name:   "int 实际为 bigint",
			std:    StdColumn{Name: "age", DataType: DataTypeInt},
			actual: ActualColumn{Name: "age", DataType: "bigint", ColumnType: "bigint"},
			want:   false,
		},
		{
			name:   "bigint 匹配",
			std:    StdColumn{Name: "uid", DataType: DataTypeBigint},
			actual: ActualColumn{Name: "uid", DataType: "bigint", ColumnType: "bigint"},
			want:   true,
		},
		{
			name:   "datetime 匹配（含小数位定义 DATA_TYPE 不变）",
			std:    StdColumn{Name: "created_at", DataType: DataTypeDatetime},
			actual: ActualColumn{Name: "created_at", DataType: "datetime", ColumnType: "datetime(3)"},
			want:   true,
		},
		{
			name:   "date 实际为 datetime",
			std:    StdColumn{Name: "birthday", DataType: DataTypeDate},
			actual: ActualColumn{Name: "birthday", DataType: "datetime", ColumnType: "datetime"},
			want:   false,
		},
		{
			name:   "text 匹配",
			std:    StdColumn{Name: "remark", DataType: DataTypeText},
			actual: ActualColumn{Name: "remark", DataType: "text", ColumnType: "text"},
			want:   true,
		},
		{
			// 未知标准类型一律不匹配
			name:   "未知标准类型",
			std:    StdColumn{Name: "ext", DataType: "json"},
			actual: ActualColumn{Name: "ext", DataType: "json", ColumnType: "json"},
			want:   false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := TypeMatched(c.std.DataType, c.std.Length, c.std.Precision, c.actual); got != c.want {
				t.Fatalf("TypeMatched(%+v, %+v) = %v, want %v", c.std, c.actual, got, c.want)
			}
		})
	}
}

// 结构比对：匹配 / 缺列 / 类型不匹配三种差异结果与整体通过标志
func TestBuildStructDiff(t *testing.T) {
	t.Run("全部匹配且列名大小写不敏感", func(t *testing.T) {
		stdCols := []*StdColumn{
			{Name: "user_name", DataType: DataTypeVarchar, Length: 50},
			{Name: "age", DataType: DataTypeInt},
		}
		actuals := []*ActualColumn{
			{Name: "USER_NAME", DataType: "varchar", ColumnType: "varchar(50)", CharLen: 50},
			{Name: "age", DataType: "int", ColumnType: "int"},
		}
		diff, pass := BuildStructDiff(stdCols, actuals)
		if !pass {
			t.Fatalf("pass = false, want true")
		}
		if len(diff) != 2 {
			t.Fatalf("diff 项数 = %d, want 2", len(diff))
		}
		for _, d := range diff {
			if d.Result != DiffResultMatch {
				t.Fatalf("column %s result = %s, want match", d.Column, d.Result)
			}
		}
	})

	t.Run("缺列 missing", func(t *testing.T) {
		stdCols := []*StdColumn{
			{Name: "user_name", DataType: DataTypeVarchar, Length: 50},
			{Name: "email", DataType: DataTypeVarchar, Length: 128},
		}
		actuals := []*ActualColumn{
			{Name: "user_name", DataType: "varchar", ColumnType: "varchar(50)", CharLen: 50},
		}
		diff, pass := BuildStructDiff(stdCols, actuals)
		if pass {
			t.Fatalf("pass = true, want false")
		}
		if len(diff) != 2 {
			t.Fatalf("diff 项数 = %d, want 2", len(diff))
		}
		if diff[1].Column != "email" || diff[1].Result != DiffResultMissing || diff[1].Actual != "" {
			t.Fatalf("缺失项 = %+v, want {email, missing, actual 空}", diff[1])
		}
		if diff[1].Expect != "varchar(128)" {
			t.Fatalf("缺失项 expect = %q, want varchar(128)", diff[1].Expect)
		}
	})

	t.Run("类型不匹配 type_mismatch", func(t *testing.T) {
		stdCols := []*StdColumn{
			{Name: "amount", DataType: DataTypeDecimal, Length: 10, Precision: 2},
		}
		actuals := []*ActualColumn{
			{Name: "amount", DataType: "int", ColumnType: "int"},
		}
		diff, pass := BuildStructDiff(stdCols, actuals)
		if pass {
			t.Fatalf("pass = true, want false")
		}
		if diff[0].Result != DiffResultTypeMismatch {
			t.Fatalf("result = %s, want type_mismatch", diff[0].Result)
		}
		if diff[0].Expect != "decimal(10,2)" || diff[0].Actual != "int" {
			t.Fatalf("差异项 = %+v, want expect decimal(10,2) actual int", diff[0])
		}
	})

	t.Run("混合差异按标准列序输出", func(t *testing.T) {
		stdCols := []*StdColumn{
			{Name: "a_col", DataType: DataTypeVarchar, Length: 50},
			{Name: "b_col", DataType: DataTypeBigint},
			{Name: "c_col", DataType: DataTypeInt},
		}
		actuals := []*ActualColumn{
			{Name: "a_col", DataType: "varchar", ColumnType: "varchar(20)", CharLen: 20},
			{Name: "c_col", DataType: "int", ColumnType: "int"},
		}
		diff, pass := BuildStructDiff(stdCols, actuals)
		if pass {
			t.Fatalf("pass = true, want false")
		}
		wantResults := []string{DiffResultTypeMismatch, DiffResultMissing, DiffResultMatch}
		for i, want := range wantResults {
			if diff[i].Result != want {
				t.Fatalf("diff[%d].result = %s, want %s", i, diff[i].Result, want)
			}
		}
		if diff[1].Column != "b_col" {
			t.Fatalf("缺失列应为 b_col, got %s", diff[1].Column)
		}
	})
}
