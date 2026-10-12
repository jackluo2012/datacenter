package tag

import (
	"datacenter/internal/types"
	"datacenter/tag/rpc/tagClient"
)

// 圈群共享映射（对齐 standard 模板：nil 安全 + Get* 逐字段 + 不映射 Beid 租户内部字段）

func groupToTypes(g *tagClient.TagGroup) *types.TagGroupInfo {
	if g == nil {
		return &types.TagGroupInfo{}
	}
	return &types.TagGroupInfo{
		Id:                g.GetId(),
		Name:              g.GetName(),
		ObjectId:          g.GetObjectId(),
		CircleRuleSql:     g.GetCircleRuleSql(),
		EstCount:          g.GetEstCount(),
		MaterializedTable: g.GetMaterializedTable(),
		LastRunAt:         g.GetLastRunAt(),
	}
}

func groupListToTypes(list []*tagClient.TagGroup) []types.TagGroupInfo {
	out := make([]types.TagGroupInfo, 0, len(list))
	for _, g := range list {
		out = append(out, *groupToTypes(g))
	}
	return out
}

func memberToTypes(m *tagClient.GroupMember) *types.GroupMember {
	if m == nil {
		return &types.GroupMember{}
	}
	return &types.GroupMember{
		ObjectPk: m.GetObjectPk(),
		TagValue: m.GetTagValue(),
	}
}
