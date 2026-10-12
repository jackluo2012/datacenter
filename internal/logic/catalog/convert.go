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

// toAsset 将 rpc 侧的 Asset 逐字段映射为网关侧的 Asset（网关契约无 beid，不透出）。
func toAsset(a *catalogClient.Asset) types.Asset {
	return types.Asset{
		Id:             a.GetId(),
		EntityId:       a.GetEntityId(),
		CategoryId:     a.GetCategoryId(),
		Name:           a.GetName(),
		EntityType:     a.GetEntityType(),
		Comment:        a.GetComment(),
		Classification: a.GetClassification(),
		AssetTagsJson:  a.GetAssetTagsJson(),
		Visibility:     a.GetVisibility(),
	}
}

// toAssetApply 将 rpc 侧的 AssetApply 逐字段映射为网关侧的 AssetApplyInfo。
func toAssetApply(a *catalogClient.AssetApply) types.AssetApplyInfo {
	return types.AssetApplyInfo{
		Id:        a.GetId(),
		AssetId:   a.GetAssetId(),
		Reason:    a.GetReason(),
		Applicant: a.GetApplicant(),
		Status:    a.GetStatus(),
		Reviewer:  a.GetReviewer(),
		Beid:      a.GetBeid(),
		CreatedAt: a.GetCreatedAt(),
	}
}
