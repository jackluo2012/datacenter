package datasource

import (
	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/internal/types"
)

// dataSourceInfoToTypes 将 rpc 的 DataSourceInfo 逐字段转换为网关类型。
// rpc 侧的 PasswordEnc/ParamsJson 为敏感或内部字段, 网关对外不回传, 故不做映射。
func dataSourceInfoToTypes(info *dataSourceClient.DataSourceInfo) *types.DataSourceInfo {
	if info == nil {
		return &types.DataSourceInfo{}
	}
	return &types.DataSourceInfo{
		Id:        info.Id,
		Name:      info.Name,
		Type:      info.Type,
		Host:      info.Host,
		Port:      info.Port,
		DbName:    info.DbName,
		Username:  info.Username,
		Status:    info.Status,
		Owner:     info.Owner,
		CreatedAt: info.CreatedAt,
		UpdatedAt: info.UpdatedAt,
	}
}
