package quality

import (
	"datacenter/internal/types"
	"datacenter/quality/rpc/qualityClient"
)

// qualityRuleToTypes 将 rpc 的 QualityRule 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func qualityRuleToTypes(rule *qualityClient.QualityRule) *types.QualityRule {
	if rule == nil {
		return &types.QualityRule{}
	}
	return &types.QualityRule{
		Id:            rule.GetId(),
		Name:          rule.GetName(),
		RuleType:      rule.GetRuleType(),
		TargetTable:   rule.GetTargetTable(),
		TargetColumn:  rule.GetTargetColumn(),
		ParamsJson:    rule.GetParamsJson(),
		ThresholdJson: rule.GetThresholdJson(),
		Status:        rule.GetStatus(),
		Owner:         rule.GetOwner(),
	}
}

// qualityJobToTypes 将 rpc 的 QualityJob 逐字段转换为网关类型。
func qualityJobToTypes(job *qualityClient.QualityJob) *types.QualityJob {
	if job == nil {
		return &types.QualityJob{}
	}
	return &types.QualityJob{
		Id:           job.GetId(),
		Name:         job.GetName(),
		DatasourceId: job.GetDatasourceId(),
		DbName:       job.GetDbName(),
		RuleIds:      job.GetRuleIds(),
		Cron:         job.GetCron(),
		LastStatus:   job.GetLastStatus(),
		LastRunAt:    job.GetLastRunAt(),
		Beid:         job.GetBeid(),
	}
}

// qualityResultToTypes 将 rpc 的 QualityResult 逐字段转换为网关类型。
func qualityResultToTypes(result *qualityClient.QualityResult) *types.QualityResult {
	if result == nil {
		return &types.QualityResult{}
	}
	return &types.QualityResult{
		Id:         result.GetId(),
		JobId:      result.GetJobId(),
		RuleId:     result.GetRuleId(),
		RuleType:   result.GetRuleType(),
		CheckValue: result.GetCheckValue(),
		PassFlag:   result.GetPassFlag(),
		DetailJson: result.GetDetailJson(),
		RunAt:      result.GetRunAt(),
	}
}
