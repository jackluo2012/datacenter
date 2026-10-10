package logic

import (
	"errors"
	"testing"
)

// data_type 白名单：7 种支持类型通过，其余拒绝
func TestIsValidDataType(t *testing.T) {
	cases := []struct {
		dataType string
		want     bool
	}{
		{"varchar", true},
		{"int", true},
		{"bigint", true},
		{"decimal", true},
		{"date", true},
		{"datetime", true},
		{"text", true},
		{"string", false},
		{"VARCHAR", false}, //大小写敏感，落库值以白名单小写为准
		{"json", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsValidDataType(c.dataType); got != c.want {
			t.Fatalf("IsValidDataType(%q) = %v, want %v", c.dataType, got, c.want)
		}
	}
}

// 词根类型白名单
func TestIsValidWordRootType(t *testing.T) {
	valid := []string{"common", "attribute", "dimension", "indicator"}
	for _, v := range valid {
		if !IsValidWordRootType(v) {
			t.Fatalf("IsValidWordRootType(%q) = false, want true", v)
		}
	}
	for _, bad := range []string{"", "Common", "other"} {
		if IsValidWordRootType(bad) {
			t.Fatalf("IsValidWordRootType(%q) = true, want false", bad)
		}
	}
}

// elements_json 校验：合法数组/空数组通过；非数组、缺 element_code、编码重复拒绝
func TestParseElementsJson(t *testing.T) {
	cases := []struct {
		name        string
		elementsRaw string
		wantCount   int
		wantErr     error
	}{
		{
			name:        "合法数组",
			elementsRaw: `[{"element_code":"user_name","required":true,"pk":false},{"element_code":"age"}]`,
			wantCount:   2,
		},
		{
			name:        "空数组合法（发布时才强制非空）",
			elementsRaw: `[]`,
			wantCount:   0,
		},
		{
			name:        "空串合法（视同未配置）",
			elementsRaw: ``,
			wantCount:   0,
		},
		{
			name:        "非 JSON",
			elementsRaw: `not-json`,
			wantErr:     ErrElementsJsonInvalid,
		},
		{
			name:        "对象不是数组",
			elementsRaw: `{"element_code":"user_name"}`,
			wantErr:     ErrElementsJsonInvalid,
		},
		{
			name:        "元素缺 element_code",
			elementsRaw: `[{"required":true}]`,
			wantErr:     ErrElementsJsonInvalid,
		},
		{
			name:        "element_code 空串",
			elementsRaw: `[{"element_code":"  "}]`,
			wantErr:     ErrElementsJsonInvalid,
		},
		{
			name:        "element_code 重复",
			elementsRaw: `[{"element_code":"user_name"},{"element_code":"user_name"}]`,
			wantErr:     ErrElementsJsonInvalid,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			elements, err := ParseElementsJson(c.elementsRaw)
			if c.wantErr != nil {
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("ParseElementsJson error = %v, want %v", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseElementsJson error: %v", err)
			}
			if len(elements) != c.wantCount {
				t.Fatalf("ParseElementsJson count = %d, want %d", len(elements), c.wantCount)
			}
		})
	}
}
