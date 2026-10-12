package security

import (
	"datacenter/internal/types"
	"datacenter/security/rpc/securityClient"
)

// secLevelToTypes 将 rpc 的 SecLevel 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func secLevelToTypes(level *securityClient.SecLevel) *types.SecLevelInfo {
	if level == nil {
		return &types.SecLevelInfo{}
	}
	return &types.SecLevelInfo{
		Id:      level.GetId(),
		Code:    level.GetCode(),
		Name:    level.GetName(),
		Comment: level.GetComment(),
	}
}

// entityTagToTypes 将 rpc 的 EntityTag 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func entityTagToTypes(tag *securityClient.EntityTag) *types.SecEntityTag {
	if tag == nil {
		return &types.SecEntityTag{}
	}
	return &types.SecEntityTag{
		Id:           tag.GetId(),
		EntityId:     tag.GetEntityId(),
		EntityName:   tag.GetEntityName(),
		LevelCode:    tag.GetLevelCode(),
		MaskRuleCode: tag.GetMaskRuleCode(),
	}
}

// maskRuleToTypes 将 rpc 的 MaskRule 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func maskRuleToTypes(rule *securityClient.MaskRule) *types.SecMaskRule {
	if rule == nil {
		return &types.SecMaskRule{}
	}
	return &types.SecMaskRule{
		Id:         rule.GetId(),
		RuleCode:   rule.GetRuleCode(),
		RuleName:   rule.GetRuleName(),
		Algorithm:  rule.GetAlgorithm(),
		ParamsJson: rule.GetParamsJson(),
	}
}

// secGrantToTypes 将 rpc 的 SecGrant 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func secGrantToTypes(grant *securityClient.SecGrant) *types.SecGrant {
	if grant == nil {
		return &types.SecGrant{}
	}
	return &types.SecGrant{
		Id:         grant.GetId(),
		Subject:    grant.GetSubject(),
		Resource:   grant.GetResource(),
		ResourceId: grant.GetResourceId(),
		LevelCode:  grant.GetLevelCode(),
		ExpireAt:   grant.GetExpireAt(),
	}
}

// normalizePage 分页默认值处理: page<=0 取 1, size<=0 取 10。
func normalizePage(page, size int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return page, size
}
