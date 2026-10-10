package tag

import (
	"datacenter/internal/types"
	"datacenter/tag/rpc/tagClient"
)

// tagObjectToTypes 将 rpc 的 TagObject 逐字段转换为网关类型。
func tagObjectToTypes(object *tagClient.TagObject) *types.TagObject {
	if object == nil {
		return &types.TagObject{}
	}
	return &types.TagObject{
		Id:         object.GetId(),
		ObjectType: object.GetObjectType(),
		Name:       object.GetName(),
		DsId:       object.GetDsId(),
		DbName:     object.GetDbName(),
		TableName:  object.GetTableName(),
		PkColumn:   object.GetPkColumn(),
		Beid:       object.GetBeid(),
	}
}

// tagDefToTypes 将 rpc 的 TagDef 逐字段转换为网关类型。
func tagDefToTypes(def *tagClient.TagDef) *types.TagDef {
	if def == nil {
		return &types.TagDef{}
	}
	return &types.TagDef{
		Id:          def.GetId(),
		ObjectId:    def.GetObjectId(),
		TagKey:      def.GetTagKey(),
		TagName:     def.GetTagName(),
		TagType:     def.GetTagType(),
		RuleSql:     def.GetRuleSql(),
		RefreshCron: def.GetRefreshCron(),
		Status:      def.GetStatus(),
		Beid:        def.GetBeid(),
	}
}
