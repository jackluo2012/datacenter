package model

import (
	"strings"
	"testing"
)

// TestHashPasswordRoundTrip bcrypt 往返：正确明文通过校验，错误明文被拒绝
func TestHashPasswordRoundTrip(t *testing.T) {
	const plain = "admin123"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword 出错: %v", err)
	}
	if !CheckPassword(hash, plain) {
		t.Fatalf("CheckPassword(正确明文) 应通过")
	}
	if CheckPassword(hash, "admin321") {
		t.Fatalf("CheckPassword(错误明文) 应拒绝")
	}
	if CheckPassword(hash, "") {
		t.Fatalf("CheckPassword(空明文) 应拒绝")
	}
}

// TestHashPasswordUniqueSalt 同明文两次哈希产出不同密文（bcrypt 随机盐），但均能校验通过
func TestHashPasswordUniqueSalt(t *testing.T) {
	h1, err := HashPassword("admin123")
	if err != nil {
		t.Fatalf("HashPassword 出错: %v", err)
	}
	h2, err := HashPassword("admin123")
	if err != nil {
		t.Fatalf("HashPassword 出错: %v", err)
	}
	if h1 == h2 {
		t.Fatalf("同明文两次哈希应不同（随机盐）")
	}
	if !CheckPassword(h1, "admin123") || !CheckPassword(h2, "admin123") {
		t.Fatalf("两次哈希均应能校验原明文")
	}
}

// TestHashPasswordFormat 哈希格式：$2a$ 前缀、总长 60，适配 sys_user.password varchar(128)
func TestHashPasswordFormat(t *testing.T) {
	hash, err := HashPassword("admin123")
	if err != nil {
		t.Fatalf("HashPassword 出错: %v", err)
	}
	if !strings.HasPrefix(hash, BcryptPrefix) {
		t.Fatalf("哈希应有 %s 前缀，实际: %q", BcryptPrefix, hash[:7])
	}
	if n := len(hash); n != 60 {
		t.Fatalf("bcrypt 哈希长度应为 60，实际 %d", n)
	}
	if len(hash) > 128 {
		t.Fatalf("哈希长度超出 password 列宽 varchar(128)")
	}
}

// TestCheckPasswordInvalidHash 非法哈希不 panic，直接返回 false（fail-closed）
func TestCheckPasswordInvalidHash(t *testing.T) {
	for _, bad := range []string{"", "not-a-hash", "$2a$10$short", "plain-password"} {
		if CheckPassword(bad, "admin123") {
			t.Fatalf("CheckPassword(非法哈希 %q) 应返回 false", bad)
		}
	}
}
