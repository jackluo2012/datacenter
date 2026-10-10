package logic

import (
	"datacenter/job/model"
	"datacenter/job/rpc/job"
)

// taskToPb 同步任务 model -> pb（ds_workflow_code/last_run_at 收窄转换，未运行为 0）
func taskToPb(t *model.SyncTask) *job.SyncTask {
	return &job.SyncTask{
		Id:             t.Id,
		Name:           t.Name,
		SourceDsId:     t.SourceDsId,
		TargetDsId:     t.TargetDsId,
		Mode:           t.Mode,
		MappingJson:    t.MappingJson,
		Cron:           t.Cron,
		DsWorkflowCode: int32(t.DsWorkflowCode),
		Status:         t.Status,
		LastRunAt:      t.LastRunAt,
		Beid:           t.Beid,
	}
}
