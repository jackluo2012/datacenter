package logic

import (
	"datacenter/tag/model"
	"datacenter/tag/rpc/tag"
)

// toPbObject model → pb（标签对象）
func toPbObject(o *model.TagObject) *tag.TagObject {
	if o == nil {
		return nil
	}
	return &tag.TagObject{
		Id:         o.Id,
		ObjectType: o.ObjectType,
		Name:       o.Name,
		DsId:       o.DsId,
		DbName:     o.DbName,
		TableName:  o.TableName,
		PkColumn:   o.PkColumn,
		Beid:       o.Beid,
	}
}

// toPbTagDef model → pb（标签定义）
func toPbTagDef(d *model.TagDef) *tag.TagDef {
	if d == nil {
		return nil
	}
	return &tag.TagDef{
		Id:          d.Id,
		ObjectId:    d.ObjectId,
		TagKey:      d.TagKey,
		TagName:     d.TagName,
		TagType:     d.TagType,
		RuleSql:     d.RuleSql,
		RefreshCron: d.RefreshCron,
		Status:      d.Status,
		Beid:        d.Beid,
	}
}
