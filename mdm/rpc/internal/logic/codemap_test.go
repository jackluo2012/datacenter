package logic

import (
	"testing"
	"time"
)

// codeTime 固定测试时间：2026-10-10 12:34:56 UTC（编码日期段断言用）
var codeTime = time.Date(2026, 10, 10, 12, 34, 56, 0, time.UTC)

// TestBuildCode 编码生成格式：prefix + date_pattern(Go 格式 20060102) + 6 位补零流水
func TestBuildCode(t *testing.T) {
	cases := []struct {
		name        string
		prefix      string
		datePattern string
		current     int64
		want        string
	}{
		{"前缀+日期+6 位流水", "CUS", "20060102", 42, "CUS20261010000042"},
		{"流水左补零到 6 位", "CUS", "20060102", 1, "CUS20261010000001"},
		{"流水超 6 位自然展开", "CUS", "20060102", 1234567, "CUS202610101234567"},
		{"无日期段", "ORD", "", 7, "ORD000007"},
		{"无前缀仅日期+流水", "", "20060102", 99, "20261010000099"},
		{"日期段到月", "CUS", "200601", 5, "CUS202610000005"},
		{"全空仅流水", "", "", 3, "000003"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := BuildCode(c.prefix, c.datePattern, codeTime, c.current)
			if got != c.want {
				t.Fatalf("BuildCode(%q,%q,%d) = %q, want %q", c.prefix, c.datePattern, c.current, got, c.want)
			}
		})
	}
}
