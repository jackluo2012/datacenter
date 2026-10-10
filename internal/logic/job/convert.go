package job

import (
	"datacenter/internal/types"
	"datacenter/job/rpc/jobClient"
)

// syncTaskToTypes 将 rpc 的 SyncTask 逐字段转换为网关类型。
func syncTaskToTypes(task *jobClient.SyncTask) *types.SyncTask {
	if task == nil {
		return &types.SyncTask{}
	}
	return &types.SyncTask{
		Id:             task.GetId(),
		Name:           task.GetName(),
		SourceDsId:     task.GetSourceDsId(),
		TargetDsId:     task.GetTargetDsId(),
		Mode:           task.GetMode(),
		MappingJson:    task.GetMappingJson(),
		Cron:           task.GetCron(),
		DsWorkflowCode: task.GetDsWorkflowCode(),
		Status:         task.GetStatus(),
		LastRunAt:      task.GetLastRunAt(),
		Beid:           task.GetBeid(),
	}
}
