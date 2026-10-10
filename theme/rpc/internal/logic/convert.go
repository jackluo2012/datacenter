package logic

import (
	"datacenter/theme/model"
	"datacenter/theme/rpc/theme"
)

// toPbDomain model → pb（主题域）
func toPbDomain(d *model.DomainRow) *theme.ThemeDomain {
	if d == nil {
		return nil
	}
	return &theme.ThemeDomain{
		Id:      d.Id,
		Code:    d.Code,
		Name:    d.Name,
		Comment: d.Comment,
		Beid:    d.Beid,
	}
}

// toPbTheme model → pb（主题）
func toPbTheme(t *model.ThemeRow) *theme.ThemeInfo {
	if t == nil {
		return nil
	}
	return &theme.ThemeInfo{
		Id:       t.Id,
		DomainId: t.DomainId,
		Code:     t.Code,
		Name:     t.Name,
		Comment:  t.Comment,
		Beid:     t.Beid,
	}
}

// toPbThemeModel model → pb（数仓模型）。pb 的 create_sql 非库列，按 fields_json
// 实时生成预览 DDL（fields 缺损时留空，不影响主体返回）
func toPbThemeModel(m *model.ThemeModelRow) *theme.ThemeModel {
	if m == nil {
		return nil
	}
	createSql := ""
	if fields, err := ParseFieldsJson(m.FieldsJson); err == nil {
		if ddl, err := buildCreateTableDDL(
			buildQualifiedTable(m.DbName, buildPhysicalTable(m.Layer, m.TableName)), fields); err == nil {
			createSql = ddl
		}
	}
	return &theme.ThemeModel{
		Id:           m.Id,
		ThemeId:      m.ThemeId,
		Layer:        m.Layer,
		TableName:    m.TableName,
		ModelName:    m.ModelName,
		FieldsJson:   m.FieldsJson,
		DimJson:      m.DimJson,
		CreateSql:    createSql,
		DatasourceId: m.DatasourceId,
		DbName:       m.DbName,
		Beid:         m.Beid,
	}
}

// toPbMv model → pb（物化视图）
func toPbMv(m *model.MvRow) *theme.MaterializedView {
	if m == nil {
		return nil
	}
	return &theme.MaterializedView{
		Id:            m.Id,
		Name:          m.Name,
		DatasourceId:  m.DatasourceId,
		DbName:        m.DbName,
		ViewName:      m.ViewName,
		SelectSql:     m.SelectSql,
		RefreshMode:   m.RefreshMode,
		LastStatus:    m.LastStatus,
		LastRefreshAt: m.LastRefreshAt,
		Beid:          m.Beid,
	}
}
