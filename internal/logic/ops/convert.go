package ops

import (
	"datacenter/internal/types"
	"datacenter/ops/rpc/opsClient"
)

// archiveTaskToTypes 将 rpc 的 ArchiveTask 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func archiveTaskToTypes(task *opsClient.ArchiveTask) *types.OpsArchiveTask {
	if task == nil {
		return &types.OpsArchiveTask{}
	}
	return &types.OpsArchiveTask{
		Id:           task.GetId(),
		Name:         task.GetName(),
		DatasourceId: task.GetDatasourceId(),
		DbName:       task.GetDbName(),
		TableName:    task.GetTableName(),
		ConditionSql: task.GetConditionSql(),
		ArchiveTable: task.GetArchiveTable(),
		DeleteSource: task.GetDeleteSource(),
		BatchSize:    task.GetBatchSize(),
		Cron:         task.GetCron(),
		LastStatus:   task.GetLastStatus(),
		LastRunAt:    task.GetLastRunAt(),
	}
}

// archiveLogToTypes 将 rpc 的 ArchiveLog 逐字段转换为网关类型。
func archiveLogToTypes(log *opsClient.ArchiveLog) *types.OpsArchiveLog {
	if log == nil {
		return &types.OpsArchiveLog{}
	}
	return &types.OpsArchiveLog{
		Id:           log.GetId(),
		TaskId:       log.GetTaskId(),
		ArchivedRows: log.GetArchivedRows(),
		DeletedRows:  log.GetDeletedRows(),
		Status:       log.GetStatus(),
		Msg:          log.GetMsg(),
		RunAt:        log.GetRunAt(),
	}
}

// loginLogToTypes 将 rpc 的 LoginLog 逐字段转换为网关类型。
func loginLogToTypes(log *opsClient.LoginLog) *types.OpsLoginLog {
	if log == nil {
		return &types.OpsLoginLog{}
	}
	return &types.OpsLoginLog{
		Id:        log.GetId(),
		Username:  log.GetUsername(),
		Uid:       log.GetUid(),
		Ip:        log.GetIp(),
		UserAgent: log.GetUserAgent(),
		Success:   log.GetSuccess(),
		Msg:       log.GetMsg(),
		CreatedAt: log.GetCreatedAt(),
	}
}

// auditLogToTypes 将 rpc 的 AuditLog 逐字段转换为网关类型。
func auditLogToTypes(log *opsClient.AuditLog) *types.OpsAuditLog {
	if log == nil {
		return &types.OpsAuditLog{}
	}
	return &types.OpsAuditLog{
		Id:         log.GetId(),
		Uid:        log.GetUid(),
		Username:   log.GetUsername(),
		Action:     log.GetAction(),
		Resource:   log.GetResource(),
		ResourceId: log.GetResourceId(),
		DetailJson: log.GetDetailJson(),
		Ip:         log.GetIp(),
		CreatedAt:  log.GetCreatedAt(),
	}
}
