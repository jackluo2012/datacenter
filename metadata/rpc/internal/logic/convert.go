package logic

import (
	"datacenter/metadata/model"
	"datacenter/metadata/rpc/metadata"
)

// entityToPb 实体 model -> pb
func entityToPb(e *model.Entity) *metadata.MetaEntity {
	return &metadata.MetaEntity{
		Id:             e.Id,
		EntityType:     e.EntityType,
		Fqn:            e.Fqn,
		Name:           e.Name,
		Comment:        e.Comment,
		Engine:         e.Engine,
		RowsCnt:        e.RowsCnt,
		PropsJson:      e.PropsJson,
		Classification: e.Classification,
		Owner:          e.Owner,
		DsId:           e.DsId,
		Beid:           e.Beid,
	}
}

// versionToPb 版本 model -> pb（published_at 转毫秒时间戳）
func versionToPb(v *model.Version) *metadata.MetaVersion {
	return &metadata.MetaVersion{
		Id:           v.Id,
		EntityId:     v.EntityId,
		Version:      v.Version,
		SnapshotJson: v.SnapshotJson,
		Operator:     v.Operator,
		PublishedAt:  v.PublishedAt.UnixMilli(),
	}
}
