package catalog

import (
	"datacenter/catalog/rpc/catalogClient"
	"datacenter/internal/types"
)

// toCategory 将 rpc 侧的 AssetCategory 逐字段映射为网关侧的 AssetCategory。
// 使用 pb 生成的 Get* 方法取值，对 nil 指针安全。
func toCategory(c *catalogClient.AssetCategory) types.AssetCategory {
	return types.AssetCategory{
		Id:       c.GetId(),
		ParentId: c.GetParentId(),
		Name:     c.GetName(),
		Beid:     c.GetBeid(),
	}
}
