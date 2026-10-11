package model

import "golang.org/x/crypto/bcrypt"

// 密码哈希：bcrypt。
// 依赖说明：golang.org/x/crypto v0.54.0 已在依赖树（go.mod indirect，go.sum 含 h1 完整校验），
// 直接 import 即可构建，无需改动 go.mod（已通过 go build ./sys/... 验证）。
const (
	// bcryptCost 哈希代价因子，与 bcrypt.DefaultCost 一致，显式声明便于后续统一调整
	bcryptCost = 10
	// BcryptPrefix bcrypt 哈希前缀（$2a$10$...，总长 60，适配 sys_user.password varchar(128)）
	BcryptPrefix = "$2a$"
)

// HashPassword 生成 bcrypt 密码哈希（每次调用自带随机盐，同明文产出不同哈希）
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword 校验明文密码与 bcrypt 哈希是否匹配（哈希格式非法时返回 false，不 panic）
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
