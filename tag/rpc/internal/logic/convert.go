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

// withObjectName 给标签定义 pb 回填所属对象名（d/o 任一为 nil 时安全跳过）
func withObjectName(d *tag.TagDef, o *model.TagObject) *tag.TagDef {
	if d != nil && o != nil {
		d.ObjectName = o.Name
	}
	return d
}

// objectNameMap 对象列表 → id 到名字的映射（纯函数，列表回显 object_name 用）
func objectNameMap(objs []*model.TagObject) map[int64]string {
	m := make(map[int64]string, len(objs))
	for _, o := range objs {
		m[o.Id] = o.Name
	}
	return m
}
