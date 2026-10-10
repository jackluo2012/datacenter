package logic

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

func strPtr(s string) *string { return &s }

// TestMaskValueMask mask 算法：保留前后缀 + 掩码填充（含中文 rune 边界与前后缀超过长度）
func TestMaskValueMask(t *testing.T) {
	cases := []struct {
		name   string
		sample string
		p      *MaskParams
		want   string
	}{
		{"手机号保留3前4后", "13812345678", &MaskParams{Prefix: 3, Suffix: 4}, "138****5678"},
		{"中文rune不乱码", "王小明", &MaskParams{Prefix: 1, Suffix: 1}, "王*明"},
		{"中文+自定义掩码字符", "银行卡号622288", &MaskParams{Prefix: 3, Suffix: 1, MaskChar: strPtr("#")}, "银行卡######8"},
		{"中文无半字符截断", "中国北京上海", &MaskParams{Prefix: 2, Suffix: 2}, "中国**上海"},
		{"前缀超过长度整值掩码", "ab", &MaskParams{Prefix: 5, Suffix: 1}, "**"},
		{"长度恰等于前后缀之和整值掩码", "abc", &MaskParams{Prefix: 1, Suffix: 2}, "***"},
		{"零保留全掩码", "13800", &MaskParams{Prefix: 0, Suffix: 0}, "*****"},
		{"空样例返回空", "", &MaskParams{Prefix: 3, Suffix: 4}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := MaskValue(AlgorithmMask, c.sample, c.p); got != c.want {
				t.Fatalf("MaskValue(mask, %q) = %q, want %q", c.sample, got, c.want)
			}
		})
	}
}

// TestMaskValueHash hash 算法：sha256 hex 前 16 位
func TestMaskValueHash(t *testing.T) {
	// 已知向量校验（sha256("abc") = ba7816bf8f01cfea...）
	if got := MaskValue(AlgorithmHash, "abc", nil); got != "ba7816bf8f01cfea" {
		t.Fatalf("MaskValue(hash, abc) = %q, want ba7816bf8f01cfea", got)
	}
	// 动态向量校验：任意值均为 sha256 hex 前 16 位
	sample := "13812345678"
	sum := sha256.Sum256([]byte(sample))
	want := hex.EncodeToString(sum[:])[:16]
	if got := MaskValue(AlgorithmHash, sample, nil); got != want {
		t.Fatalf("MaskValue(hash, %q) = %q, want %q", sample, got, want)
	}
}

// TestMaskValueReplaceAndNull replace/null 算法
func TestMaskValueReplaceAndNull(t *testing.T) {
	// replace：整值替换
	if got := MaskValue(AlgorithmReplace, "13812345678", &MaskParams{Replace: strPtr("HIDDEN")}); got != "HIDDEN" {
		t.Fatalf("MaskValue(replace) = %q, want HIDDEN", got)
	}
	// replace：允许整值替换为空串（键存在即可）
	if got := MaskValue(AlgorithmReplace, "13812345678", &MaskParams{Replace: strPtr("")}); got != "" {
		t.Fatalf("MaskValue(replace 空串) = %q, want 空串", got)
	}
	// null：返回空串
	if got := MaskValue(AlgorithmNull, "secret", nil); got != "" {
		t.Fatalf("MaskValue(null) = %q, want 空串", got)
	}
	// 未知算法 fail-closed：全掩码兜底
	if got := MaskValue("unknown", "abc", nil); got != "***" {
		t.Fatalf("MaskValue(unknown) = %q, want ***", got)
	}
}

// TestParseMaskParams params_json 按算法结构校验
func TestParseMaskParams(t *testing.T) {
	// mask：合法参数
	p, err := ParseMaskParams(AlgorithmMask, `{"prefix":3,"suffix":2}`)
	if err != nil || p.Prefix != 3 || p.Suffix != 2 {
		t.Fatalf("ParseMaskParams(mask 合法) = %+v, %v", p, err)
	}
	if maskCharOf(p) != "*" {
		t.Fatalf("mask_char 缺省应为 *，got %q", maskCharOf(p))
	}
	// mask：缺 params_json 拒绝
	if _, err := ParseMaskParams(AlgorithmMask, ""); !errors.Is(err, ErrMaskParamsInvalid) {
		t.Fatalf("ParseMaskParams(mask 缺参) err = %v, want ErrMaskParamsInvalid", err)
	}
	// mask：负数拒绝
	if _, err := ParseMaskParams(AlgorithmMask, `{"prefix":-1,"suffix":2}`); !errors.Is(err, ErrMaskParamsInvalid) {
		t.Fatalf("ParseMaskParams(mask 负数) err = %v, want ErrMaskParamsInvalid", err)
	}
	// mask：掩码字符多 rune 拒绝
	if _, err := ParseMaskParams(AlgorithmMask, `{"prefix":1,"suffix":1,"mask_char":"**"}`); !errors.Is(err, ErrMaskParamsInvalid) {
		t.Fatalf("ParseMaskParams(mask 多rune掩码) err = %v, want ErrMaskParamsInvalid", err)
	}
	// mask：单个中文 rune 掩码合法
	if _, err := ParseMaskParams(AlgorithmMask, `{"prefix":1,"suffix":1,"mask_char":"遮"}`); err != nil {
		t.Fatalf("ParseMaskParams(mask 中文单rune掩码) err = %v", err)
	}
	// replace：replace 键存在即可（允许空串）
	p, err = ParseMaskParams(AlgorithmReplace, `{"replace":""}`)
	if err != nil || p.Replace == nil || *p.Replace != "" {
		t.Fatalf("ParseMaskParams(replace 空串) = %+v, %v", p, err)
	}
	// replace：缺 replace 键拒绝
	if _, err := ParseMaskParams(AlgorithmReplace, `{"prefix":1}`); !errors.Is(err, ErrMaskParamsInvalid) {
		t.Fatalf("ParseMaskParams(replace 缺键) err = %v, want ErrMaskParamsInvalid", err)
	}
	// replace：缺 params_json 拒绝
	if _, err := ParseMaskParams(AlgorithmReplace, ""); !errors.Is(err, ErrMaskParamsInvalid) {
		t.Fatalf("ParseMaskParams(replace 缺参) err = %v, want ErrMaskParamsInvalid", err)
	}
	// hash/null：无参数要求
	if _, err := ParseMaskParams(AlgorithmHash, ""); err != nil {
		t.Fatalf("ParseMaskParams(hash 空参) err = %v", err)
	}
	if _, err := ParseMaskParams(AlgorithmNull, `{"whatever":1}`); err != nil {
		t.Fatalf("ParseMaskParams(null 任意参) err = %v", err)
	}
}

// TestIsValidAlgorithm 算法白名单
func TestIsValidAlgorithm(t *testing.T) {
	for _, a := range []string{AlgorithmMask, AlgorithmHash, AlgorithmReplace, AlgorithmNull} {
		if !IsValidAlgorithm(a) {
			t.Fatalf("IsValidAlgorithm(%q) = false, want true", a)
		}
	}
	for _, a := range []string{"", "MASK", "aes", "truncate"} {
		if IsValidAlgorithm(a) {
			t.Fatalf("IsValidAlgorithm(%q) = true, want false", a)
		}
	}
}
