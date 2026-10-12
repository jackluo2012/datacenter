package theme

import (
	"datacenter/internal/types"
	"datacenter/theme/rpc/themeClient"
)

// themeDomainToTypes 将 rpc 的 ThemeDomain 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func themeDomainToTypes(domain *themeClient.ThemeDomain) *types.ThemeDomainInfo {
	if domain == nil {
		return &types.ThemeDomainInfo{}
	}
	return &types.ThemeDomainInfo{
		Id:      domain.GetId(),
		Code:    domain.GetCode(),
		Name:    domain.GetName(),
		Comment: domain.GetComment(),
	}
}

// themeToTypes 将 rpc 的 ThemeInfo 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func themeToTypes(theme *themeClient.ThemeInfo) *types.ThemeInfo {
	if theme == nil {
		return &types.ThemeInfo{}
	}
	return &types.ThemeInfo{
		Id:       theme.GetId(),
		DomainId: theme.GetDomainId(),
		Code:     theme.GetCode(),
		Name:     theme.GetName(),
		Comment:  theme.GetComment(),
	}
}

// themeModelToTypes 将 rpc 的 ThemeModel 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射;
// create_sql 可能在 save 阶段未生成(deploy 时才生成), 为空串时原样透出。
func themeModelToTypes(model *themeClient.ThemeModel) *types.ThemeModelInfo {
	if model == nil {
		return &types.ThemeModelInfo{}
	}
	return &types.ThemeModelInfo{
		Id:           model.GetId(),
		ThemeId:      model.GetThemeId(),
		Layer:        model.GetLayer(),
		TableName:    model.GetTableName(),
		ModelName:    model.GetModelName(),
		FieldsJson:   model.GetFieldsJson(),
		DimJson:      model.GetDimJson(),
		CreateSql:    model.GetCreateSql(),
		DatasourceId: model.GetDatasourceId(),
		DbName:       model.GetDbName(),
	}
}

// mvToTypes 将 rpc 的 MaterializedView 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射;
// last_status/last_refresh_at 由刷新动作回写, 新建记录为空串/0 原样透出。
func mvToTypes(mv *themeClient.MaterializedView) *types.ThemeMvInfo {
	if mv == nil {
		return &types.ThemeMvInfo{}
	}
	return &types.ThemeMvInfo{
		Id:            mv.GetId(),
		Name:          mv.GetName(),
		DatasourceId:  mv.GetDatasourceId(),
		DbName:        mv.GetDbName(),
		ViewName:      mv.GetViewName(),
		SelectSql:     mv.GetSelectSql(),
		RefreshMode:   mv.GetRefreshMode(),
		LastStatus:    mv.GetLastStatus(),
		LastRefreshAt: mv.GetLastRefreshAt(),
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
