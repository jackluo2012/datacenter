package logic

import (
	"bytes"
	"strings"
	"testing"
)

// testKey 32 字节测试密钥(与 example yaml 一致)
const testKey = "0123456789abcdef0123456789abcdef"

// wrongKey 另一把合法长度但内容不同的密钥
const wrongKey = "ffffffffffffffffffffffffffffffff"

// TestEncryptDecryptRoundTrip 加密→解密往返一致
func TestEncryptDecryptRoundTrip(t *testing.T) {
	plainList := []string{
		"",                       // 空密码
		"admin123",               // 常规密码
		"p@ss w0rd!#$%^&*()",     // 特殊字符
		"中文密码测试123",              // 中文
		strings.Repeat("x", 200), // 超长密码
	}
	for _, plain := range plainList {
		enc, err := EncryptPassword(testKey, plain)
		if err != nil {
			t.Fatalf("加密 %q 失败: %v", plain, err)
		}
		// 密文长度必须大于 nonce 长度(携带 GCM tag)
		if len(enc) <= aesGcmNonceLen {
			t.Fatalf("密文长度 %d 不合法", len(enc))
		}
		// 非空明文不允许在密文中出现
		if plain != "" && bytes.Contains(enc, []byte(plain)) {
			t.Fatalf("密文中出现了明文")
		}
		got, err := DecryptPassword(testKey, enc)
		if err != nil {
			t.Fatalf("解密 %q 失败: %v", plain, err)
		}
		if got != plain {
			t.Fatalf("往返不一致, 期望 %q, 实际 %q", plain, got)
		}
	}
}

// TestEncryptRandomNonce 相同明文两次加密结果不同(随机 nonce 前置)
func TestEncryptRandomNonce(t *testing.T) {
	a, err := EncryptPassword(testKey, "admin123")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	b, err := EncryptPassword(testKey, "admin123")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if bytes.Equal(a, b) {
		t.Fatal("两次加密结果相同, nonce 未随机化")
	}
	// nonce 前置: 前 12 字节应不同
	if bytes.Equal(a[:aesGcmNonceLen], b[:aesGcmNonceLen]) {
		t.Fatal("两次加密的 nonce 相同")
	}
}

// TestDecryptWithWrongKey 错误密钥解密必须失败
func TestDecryptWithWrongKey(t *testing.T) {
	enc, err := EncryptPassword(testKey, "admin123")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if _, err = DecryptPassword(wrongKey, enc); err == nil {
		t.Fatal("错误密钥解密应当失败")
	}
}

// TestDecryptWithTamperedCiphertext 篡改密文(nonce 或密文体)解密必须失败
func TestDecryptWithTamperedCiphertext(t *testing.T) {
	enc, err := EncryptPassword(testKey, "admin123")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	// 篡改密文体最后一个字节(GCM tag 校验应失败)
	tampered := bytes.Clone(enc)
	tampered[len(tampered)-1] ^= 0xff
	if _, err = DecryptPassword(testKey, tampered); err == nil {
		t.Fatal("篡改密文后解密应当失败")
	}
	// 篡改 nonce 部分
	tamperedNonce := bytes.Clone(enc)
	tamperedNonce[0] ^= 0xff
	if _, err = DecryptPassword(testKey, tamperedNonce); err == nil {
		t.Fatal("篡改 nonce 后解密应当失败")
	}
	// 截断密文
	if _, err = DecryptPassword(testKey, enc[:aesGcmNonceLen]); err == nil {
		t.Fatal("截断密文解密应当失败")
	}
}

// TestInvalidKeyLength 非 32 字节密钥必须报错
func TestInvalidKeyLength(t *testing.T) {
	if _, err := EncryptPassword("short-key", "admin123"); err == nil {
		t.Fatal("非 32 字节密钥加密应当失败")
	}
	enc, err := EncryptPassword(testKey, "admin123")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if _, err = DecryptPassword("short-key", enc); err == nil {
		t.Fatal("非 32 字节密钥解密应当失败")
	}
}

// TestDecryptEmptyCiphertext 空密文(未设置密码)解密返回空串且不报错
func TestDecryptEmptyCiphertext(t *testing.T) {
	got, err := DecryptPassword(testKey, nil)
	if err != nil {
		t.Fatalf("空密文解密不应报错: %v", err)
	}
	if got != "" {
		t.Fatalf("空密文解密应返回空串, 实际 %q", got)
	}
}

// TestBuildMysqlDsn mysql 连接串拼接格式
func TestBuildMysqlDsn(t *testing.T) {
	want := "user:pass@tcp(127.0.0.1:3306)/datacenter?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	got := BuildMysqlDsn("127.0.0.1", 3306, "datacenter", "user", "pass")
	if got != want {
		t.Fatalf("DSN 不一致, 期望 %s, 实际 %s", want, got)
	}
}
