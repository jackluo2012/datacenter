package logic

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"datacenter/datasvc/rpc/internal/svc"
)

// signToleranceSeconds 签名时间戳允许的服务器时间偏差（秒），超过即拒绝
const signToleranceSeconds int64 = 300

// aesKey 规范化 AES 密钥：32 字节直接使用，否则用 sha256 派生 32 字节（AES-256）
func aesKey(key string) []byte {
	if len(key) == 32 {
		return []byte(key)
	}
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}

// aesGcmEncrypt AES-GCM 加密，输出 nonce||密文 的字节串（存 varbinary 列）
func aesGcmEncrypt(key, plaintext string) ([]byte, error) {
	block, err := aes.NewCipher(aesKey(key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

// aesGcmDecrypt AES-GCM 解密（输入 nonce||密文）
func aesGcmDecrypt(key string, data []byte) (string, error) {
	block, err := aes.NewCipher(aesKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("密文长度非法")
	}
	plain, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// randomHex 生成 n 字节加密随机数的 hex 字符串（2n 个字符）
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// genSign 计算执行签名：hex(hmac-sha256(key=app_secret, msg=app_key+ts))
func genSign(appKey, ts, appSecret string) string {
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(appKey + ts))
	return hex.EncodeToString(mac.Sum(nil))
}

// verifySign 校验执行请求签名（纯函数，now 由调用方注入便于测试）：
// 1) ts 必须是合法秒级时间戳；2) 与服务器时间偏差不得超过 300 秒；
// 3) sign 必须与 hex(hmac-sha256(app_key+ts)) 常量时间比对一致。
func verifySign(appKey, ts, sign, appSecret string, now time.Time) error {
	tsv, err := strconv.ParseInt(strings.TrimSpace(ts), 10, 64)
	if err != nil {
		return fmt.Errorf("%w: ts 非法", ErrSignInvalid)
	}
	diff := now.Unix() - tsv
	if diff < 0 {
		diff = -diff
	}
	if diff > signToleranceSeconds {
		return fmt.Errorf("%w: ts 与服务器时间偏差超过 %d 秒", ErrSignInvalid, signToleranceSeconds)
	}
	expect, err := hex.DecodeString(genSign(appKey, ts, appSecret))
	if err != nil {
		return fmt.Errorf("%w: 签名计算失败", ErrSignInvalid)
	}
	got, err := hex.DecodeString(strings.TrimSpace(sign))
	if err != nil {
		return fmt.Errorf("%w: sign 格式非法", ErrSignInvalid)
	}
	if !hmac.Equal(expect, got) {
		return fmt.Errorf("%w: 签名不匹配", ErrSignInvalid)
	}
	return nil
}

// ipAllowed 校验来源 IP：白名单为空表示不限；非空时 ip 必须精确命中其中一项（忽略空白项）
func ipAllowed(ip, whitelist string) bool {
	whitelist = strings.TrimSpace(whitelist)
	if whitelist == "" {
		return true
	}
	for _, item := range strings.Split(whitelist, ",") {
		if strings.TrimSpace(item) == ip {
			return true
		}
	}
	return false
}

// encryptAppSecret 用 config.EncryptKey 加密应用密钥
func encryptAppSecret(svcCtx *svc.ServiceContext, secret string) ([]byte, error) {
	return aesGcmEncrypt(svcCtx.Config.EncryptKey, secret)
}

// decryptAppSecret 用 config.EncryptKey 解密应用密钥
func decryptAppSecret(svcCtx *svc.ServiceContext, data []byte) (string, error) {
	return aesGcmDecrypt(svcCtx.Config.EncryptKey, data)
}
