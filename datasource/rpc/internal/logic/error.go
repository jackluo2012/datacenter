package logic

import "errors"

// 业务错误定义(对齐 votes 服务风格), 交由 xzrpc 以非零 code 返回
var (
	ErrNotFound          = errors.New("数据源不存在")
	ErrNameDuplicate     = errors.New("数据源名称已存在")
	ErrNameEmpty         = errors.New("数据源名称不能为空")
	ErrTypeInvalid       = errors.New("数据源类型不支持, 仅支持: mysql/redis/es/http/kafka")
	ErrEncryptKeyInvalid = errors.New("EncryptKey 配置不合法, 必须为 32 字节")
	ErrEncryptFail       = errors.New("密码加密失败")
	ErrDecryptFail       = errors.New("密码解密失败, 密钥不匹配或密文损坏")
	ErrCreateFail        = errors.New("创建数据源失败")
	ErrUpdateFail        = errors.New("更新数据源失败")
	ErrDeleteFail        = errors.New("删除数据源失败")
	ErrQueryFail         = errors.New("查询数据源失败")
)

// datasourceTypes 支持的数据源类型白名单
var datasourceTypes = map[string]struct{}{
	"mysql": {},
	"redis": {},
	"es":    {},
	"http":  {},
	"kafka": {},
}

// validDataSourceType 校验数据源类型是否在白名单内
func validDataSourceType(typ string) bool {
	_, ok := datasourceTypes[typ]
	return ok
}
