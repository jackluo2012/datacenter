package logic

import (
	"datacenter/quality/model"
	"datacenter/quality/rpc/quality"
)

// ruleToPb 规则 model -> pb（status 为 int32 收窄转换）
func ruleToPb(r *model.Rule) *quality.QualityRule {
	return &quality.QualityRule{
		Id:            r.Id,
		Name:          r.Name,
		RuleType:      r.RuleType,
		TargetTable:   r.TargetTable,
		TargetColumn:  r.TargetColumn,
		ParamsJson:    r.ParamsJson,
		ThresholdJson: r.ThresholdJson,
		Status:        int32(r.Status),
		Owner:         r.Owner,
		Beid:          r.Beid,
	}
}

// jobToPb 任务 model -> pb（last_run_at 为毫秒时间戳，未运行为 0）
func jobToPb(j *model.Job) *quality.QualityJob {
	return &quality.QualityJob{
		Id:           j.Id,
		Name:         j.Name,
		DatasourceId: j.DatasourceId,
		DbName:       j.DbName,
		RuleIds:      j.RuleIds,
		Cron:         j.Cron,
		LastStatus:   j.LastStatus,
		LastRunAt:    j.LastRunAt,
		Beid:         j.Beid,
	}
}

// resultToPb 结果 model -> pb（pass_flag tinyint -> bool，run_at 为毫秒时间戳）
func resultToPb(r *model.Result) *quality.QualityResult {
	return &quality.QualityResult{
		Id:         r.Id,
		JobId:      r.JobId,
		RuleId:     r.RuleId,
		RuleType:   r.RuleType,
		CheckValue: r.CheckValue,
		PassFlag:   r.PassFlag == 1,
		DetailJson: r.DetailJson,
		RunAt:      r.RunAt,
	}
}
