package ide

import (
	"datacenter/ide/rpc/ideClient"
	"datacenter/internal/types"
)

// scriptToTypes 将 rpc 的 Script 逐字段转换为网关类型。
func scriptToTypes(script *ideClient.Script) *types.Script {
	if script == nil {
		return &types.Script{}
	}
	return &types.Script{
		Id:           script.GetId(),
		Name:         script.GetName(),
		SqlText:      script.GetSqlText(),
		Version:      script.GetVersion(),
		DatasourceId: script.GetDatasourceId(),
		DbName:       script.GetDbName(),
		LastRunAt:    script.GetLastRunAt(),
		Beid:         script.GetBeid(),
	}
}
