package metadata

import (
	"datacenter/internal/types"
	"datacenter/metadata/rpc/metadataClient"
)

// entityToTypes 将 rpc 的 MetaEntity 逐字段转换为网关类型
func entityToTypes(entity *metadataClient.MetaEntity) *types.MetaEntity {
	if entity == nil {
		return &types.MetaEntity{}
	}
	return &types.MetaEntity{
		Id:             entity.Id,
		EntityType:     entity.EntityType,
		Fqn:            entity.Fqn,
		Name:           entity.Name,
		Comment:        entity.Comment,
		Engine:         entity.Engine,
		RowsCnt:        entity.RowsCnt,
		PropsJson:      entity.PropsJson,
		Classification: entity.Classification,
		Owner:          entity.Owner,
		DsId:           entity.DsId,
		Beid:           entity.Beid,
	}
}

// versionToTypes 将 rpc 的 MetaVersion 逐字段转换为网关类型
func versionToTypes(version *metadataClient.MetaVersion) *types.MetaVersion {
	if version == nil {
		return &types.MetaVersion{}
	}
	return &types.MetaVersion{
		Id:           version.Id,
		EntityId:     version.EntityId,
		Version:      version.Version,
		SnapshotJson: version.SnapshotJson,
		Operator:     version.Operator,
		PublishedAt:  version.PublishedAt,
	}
}

// lineageNodesToTypes 将 rpc 的血缘节点列表逐字段转换为网关类型
func lineageNodesToTypes(nodes []*metadataClient.LineageNode) []types.LineageNode {
	list := make([]types.LineageNode, 0, len(nodes))
	for _, node := range nodes {
		list = append(list, types.LineageNode{
			Id:   node.GetId(),
			Fqn:  node.GetFqn(),
			Name: node.GetName(),
		})
	}
	return list
}

// lineageEdgesToTypes 将 rpc 的血缘边列表逐字段转换为网关类型
func lineageEdgesToTypes(edges []*metadataClient.LineageEdge) []types.LineageEdge {
	list := make([]types.LineageEdge, 0, len(edges))
	for _, edge := range edges {
		list = append(list, types.LineageEdge{
			FromId:    edge.GetFromId(),
			ToId:      edge.GetToId(),
			PropsJson: edge.GetPropsJson(),
		})
	}
	return list
}
