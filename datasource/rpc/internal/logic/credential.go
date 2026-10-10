package logic

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

const (
	// aesGcmNonceLen AES-GCM 标准 nonce 长度(12 字节)
	aesGcmNonceLen = 12
	// aesGcmKeyLen 配置密钥要求长度(32 字节, 即 AES-256)
	aesGcmKeyLen = 32
)

// newGCM 校验密钥长度并构造 AES-GCM 实例
func newGCM(key string) (cipher.AEAD, error) {
	if len(key) != aesGcmKeyLen {
		return nil, ErrEncryptKeyInvalid
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("构造 AES cipher 失败: %w", err)
	}
	return cipher.NewGCM(block)
}

// EncryptPassword AES-GCM 加密: 随机 nonce(12 字节) 前置拼接密文后整体落库(varbinary 列)
// 存储布局: nonce(12B) + ciphertext(含 GCM tag)
func EncryptPassword(key, plaintext string) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesGcmNonceLen)
	if _, err = rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptFail, err)
	}
	// Seal 输出 = nonce 拼接密文, 解密时按前 12 字节切分
	return gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

// DecryptPassword 解密 nonce 前置拼接的 AES-GCM 密文; 空密文(未设置密码)直接返回空串
func DecryptPassword(key string, ciphertext []byte) (string, error) {
	if len(ciphertext) == 0 {
		return "", nil
	}
	if len(ciphertext) <= aesGcmNonceLen {
		return "", errors.New("密文长度不合法")
	}
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	plaintext, err := gcm.Open(nil, ciphertext[:aesGcmNonceLen], ciphertext[aesGcmNonceLen:], nil)
	if err != nil {
		return "", ErrDecryptFail
	}
	return string(plaintext), nil
}

// BuildMysqlDsn 拼接 mysql 连接串(用户名/密码不转义, 与 go-sql-driver DSN 格式一致)
func BuildMysqlDsn(host string, port int32, dbName, username, password string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Asia%%2FShanghai",
		username, password, host, port, dbName)
}
