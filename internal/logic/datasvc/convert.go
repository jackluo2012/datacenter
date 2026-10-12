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

// callLogToTypes 将 rpc 侧的 ApiCallLog 逐字段映射为网关侧的 DsApiCallLog。
// 两边字段一一对应(id/api_id/app_id/path/status/error_code/rows_cnt/cost_ms/
// cache_hit/ip/created_at), 使用 pb 生成的 Get* 方法取值, 对 nil 指针安全。
func callLogToTypes(log *dataSvcClient.ApiCallLog) types.DsApiCallLog {
	if log == nil {
		return types.DsApiCallLog{}
	}
	return types.DsApiCallLog{
		Id:        log.GetId(),
		ApiId:     log.GetApiId(),
		AppId:     log.GetAppId(),
		Path:      log.GetPath(),
		Status:    log.GetStatus(),
		ErrorCode: log.GetErrorCode(),
		RowsCnt:   log.GetRowsCnt(),
		CostMs:    log.GetCostMs(),
		CacheHit:  log.GetCacheHit(),
		Ip:        log.GetIp(),
		CreatedAt: log.GetCreatedAt(),
	}
}

// statsOverallToTypes 将 rpc 侧的 ApiCallStats 逐字段映射为网关侧的 DsStatsOverall。
func statsOverallToTypes(stats *dataSvcClient.ApiCallStats) types.DsStatsOverall {
	if stats == nil {
		return types.DsStatsOverall{}
	}
	return types.DsStatsOverall{
		TotalCalls:  stats.GetTotalCalls(),
		OkCalls:     stats.GetOkCalls(),
		ErrorCalls:  stats.GetErrorCalls(),
		DeniedCalls: stats.GetDeniedCalls(),
		SuccessRate: stats.GetSuccessRate(),
		AvgCostMs:   stats.GetAvgCostMs(),
		CacheHits:   stats.GetCacheHits(),
	}
}

// topApiToTypes 将 rpc 侧的 TopApi 逐字段映射为网关侧的 DsTopApi。
func topApiToTypes(top *dataSvcClient.TopApi) types.DsTopApi {
	if top == nil {
		return types.DsTopApi{}
	}
	return types.DsTopApi{
		ApiId:     top.GetApiId(),
		Path:      top.GetPath(),
		Calls:     top.GetCalls(),
		AvgCostMs: top.GetAvgCostMs(),
	}
}
