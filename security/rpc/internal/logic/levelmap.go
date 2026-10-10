package logic

import (
	"fmt"
	"strconv"
)

// LevelRank 密级数字化（纯函数）：L1→1 … L9→9；
// 非法/未知编码返回 0（fail-closed：参与鉴权比较时一律拒绝）
func LevelRank(code string) int {
	if !ValidLevelCode(code) {
		return 0
	}
	rank, err := strconv.Atoi(code[1:])
	if err != nil {
		return 0
	}
	return rank
}

// DecidePerm 密级比较（纯函数）：数据密级 ≤ 许可密级才放行，
// 任一侧编码未知（rank=0）一律拒绝。返回 (是否放行, 原因说明)
func DecidePerm(grantLevel, dataLevel string) (bool, string) {
	dataRank := LevelRank(dataLevel)
	if dataRank == 0 {
		return false, "数据密级非法：" + dataLevel
	}
	grantRank := LevelRank(grantLevel)
	if grantRank == 0 {
		return false, "许可密级非法：" + grantLevel
	}
	if dataRank <= grantRank {
		return true, "数据密级不超过许可密级，允许访问"
	}
	return false, fmt.Sprintf("数据密级 %s 高于许可密级 %s，拒绝访问", dataLevel, grantLevel)
}

// GrantExpired 授权是否已过期（纯函数）：expire_at=0 表示永久不过期，等于过期时刻视为已过期
func GrantExpired(expireAt, now int64) bool {
	return expireAt > 0 && expireAt <= now
}
