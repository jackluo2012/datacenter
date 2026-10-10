package datasvc

import (
	"datacenter/datasvc/rpc/dataSvcClient"
	"datacenter/internal/types"
)

// apiToInfo 将 rpc 侧的 Api 逐字段映射为网关侧的 ApiInfo。
// 网关契约（types.ApiInfo）没有 beid / resp_sample_json 字段，故不透出。
// 使用 pb 生成的 Get* 方法取值，对 nil 指针安全。
func apiToInfo(api *dataSvcClient.Api) types.ApiInfo {
	return types.ApiInfo{
		Id:           api.GetId(),
		Name:         api.GetName(),
		Path:         api.GetPath(),
		Mode:         api.GetMode(),
		DatasourceId: api.GetDatasourceId(),
		SqlText:      api.GetSqlText(),
		ParamsJson:   api.GetParamsJson(),
		CacheTtl:     api.GetCacheTtl(),
		Status:       api.GetStatus(),
		Owner:        api.GetOwner(),
	}
}
