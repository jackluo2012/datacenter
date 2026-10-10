package logic

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

// mustSign 按契约计算签名：hex(hmac-sha256(key=app_secret, msg=app_key+ts))
func mustSign(t *testing.T, appKey, ts, secret string) string {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(appKey + ts))
	return hex.EncodeToString(mac.Sum(nil))
}

// 签名校验：正确签名 / 过期 ts / 错误密钥 / 非法 ts / 非法 sign
func TestVerifySign(t *testing.T) {
	now := time.Unix(1700000000, 0)
	const (
		appKey = "app_0123456789abcdef"
		secret = "0123456789abcdef0123456789abcdef"
	)
	ts := "1700000000"

	// 正确签名
	if err := verifySign(appKey, ts, mustSign(t, appKey, ts, secret), secret, now); err != nil {
		t.Fatalf("正确签名不应报错: %v", err)
	}
	// 错误密钥
	if err := verifySign(appKey, ts, mustSign(t, appKey, ts, "wrong-secret"), secret, now); !errors.Is(err, ErrSignInvalid) {
		t.Fatalf("错误密钥应返回 ErrSignInvalid, got: %v", err)
	}
	// 过期 ts（早 400 秒，超过 300 秒窗口）
	oldTs := "1699999600"
	if err := verifySign(appKey, oldTs, mustSign(t, appKey, oldTs, secret), secret, now); !errors.Is(err, ErrSignInvalid) {
		t.Fatalf("过期 ts 应返回 ErrSignInvalid, got: %v", err)
	}
	// 未来 ts 超窗（晚 400 秒）
	futureTs := "1700000400"
	if err := verifySign(appKey, futureTs, mustSign(t, appKey, futureTs, secret), secret, now); !errors.Is(err, ErrSignInvalid) {
		t.Fatalf("未来 ts 超窗应返回 ErrSignInvalid, got: %v", err)
	}
	// 恰好 300 秒边界，应放行
	boundaryTs := "1699999700"
	if err := verifySign(appKey, boundaryTs, mustSign(t, appKey, boundaryTs, secret), secret, now); err != nil {
		t.Fatalf("边界 ts（300s）不应报错: %v", err)
	}
	// ts 非数字
	if err := verifySign(appKey, "abc", mustSign(t, appKey, "abc", secret), secret, now); !errors.Is(err, ErrSignInvalid) {
		t.Fatalf("非法 ts 应返回 ErrSignInvalid, got: %v", err)
	}
	// sign 非 hex
	if err := verifySign(appKey, ts, "zz-not-hex", secret, now); !errors.Is(err, ErrSignInvalid) {
		t.Fatalf("非法 sign 应返回 ErrSignInvalid, got: %v", err)
	}
	// sign 长度不足
	if err := verifySign(appKey, ts, "abcd", secret, now); !errors.Is(err, ErrSignInvalid) {
		t.Fatalf("长度不足 sign 应返回 ErrSignInvalid, got: %v", err)
	}
}

// AES-GCM 加解密回环：加密后可解出明文，篡改密文解密失败
func TestAesGcmRoundTrip(t *testing.T) {
	const key = "0123456789abcdef0123456789abcdef" // 32 字节
	plain := "app_secret_plain_text"
	enc, err := aesGcmEncrypt(key, plain)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	got, err := aesGcmDecrypt(key, enc)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if got != plain {
		t.Fatalf("回环不一致: want %q got %q", plain, got)
	}
	// 非对称密钥（非 32 字节）走 sha256 派生，也应回环一致
	enc2, _ := aesGcmEncrypt("short-key", plain)
	got2, err := aesGcmDecrypt("short-key", enc2)
	if err != nil || got2 != plain {
		t.Fatalf("派生密钥回环失败: %v %q", err, got2)
	}
	// 篡改密文应解密失败
	if len(enc) > 0 {
		bad := append([]byte{}, enc...)
		bad[len(bad)-1] ^= 0xff
		if _, err := aesGcmDecrypt(key, bad); err == nil {
			t.Fatal("篡改后的密文不应解密成功")
		}
	}
	// 密文过短
	if _, err := aesGcmDecrypt(key, []byte{1, 2}); err == nil {
		t.Fatal("过短密文不应解密成功")
	}
}

// IP 白名单：空=不限；非空需精确命中（忽略空白项）
func TestIpAllowed(t *testing.T) {
	cases := []struct {
		ip        string
		whitelist string
		want      bool
	}{
		{"10.0.0.1", "", true},                         //空白名单=不限
		{"10.0.0.1", "10.0.0.1,192.168.1.1", true},     //命中
		{"10.0.0.1", " 10.0.0.1 , 192.168.1.1 ", true}, //带空格命中
		{"10.0.0.2", "10.0.0.1,192.168.1.1", false},    //未命中
		{"", "10.0.0.1", false},                        //空 ip 不放行
	}
	for _, c := range cases {
		if got := ipAllowed(c.ip, c.whitelist); got != c.want {
			t.Fatalf("ipAllowed(%q, %q) = %v, want %v", c.ip, c.whitelist, got, c.want)
		}
	}
}

// 随机串：长度与唯一性
func TestRandomHex(t *testing.T) {
	a, err := randomHex(8)
	if err != nil || len(a) != 16 {
		t.Fatalf("randomHex(8) 长度应为 16, got %q err %v", a, err)
	}
	b, _ := randomHex(16)
	if len(b) != 32 {
		t.Fatalf("randomHex(16) 长度应为 32, got %q", b)
	}
	if a == b[:16] {
		t.Fatal("随机串不应重复")
	}
}
