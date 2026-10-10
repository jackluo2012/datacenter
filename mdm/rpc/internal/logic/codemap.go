package logic

import (
	"fmt"
	"time"
)

// codemap.go 业务编码拼装（纯函数）：prefix + date_pattern(Go 时间格式) + 6 位补零流水

// codeSeqWidth 流水号补零位数（不足左补 0，超过按实际位数展开）
const codeSeqWidth = 6

// BuildCode 拼接业务编码：prefix + date_pattern（Go 格式如 20060102，空为无日期段）+ 补零流水。
// 例：prefix=CUS、date_pattern=20060102、current=42 → "CUS20261010000042"
func BuildCode(prefix, datePattern string, now time.Time, current int64) string {
	code := prefix
	if datePattern != "" {
		code += now.Format(datePattern)
	}
	return fmt.Sprintf("%s%0*d", code, codeSeqWidth, current)
}
