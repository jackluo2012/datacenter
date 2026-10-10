package mdm

import (
	"datacenter/internal/types"
	"datacenter/mdm/rpc/mdmClient"
)

// mdmModelToTypes 将 rpc 的 MdmModel 逐字段转换为网关类型。
func mdmModelToTypes(model *mdmClient.MdmModel) *types.MdmModelInfo {
	if model == nil {
		return &types.MdmModelInfo{}
	}
	return &types.MdmModelInfo{
		Id:           model.GetId(),
		ModelCode:    model.GetModelCode(),
		ModelName:    model.GetModelName(),
		TableName:    model.GetTableName(),
		FieldsJson:   model.GetFieldsJson(),
		CodeStrategy: model.GetCodeStrategy(),
		Beid:         model.GetBeid(),
	}
}

// mdmRecordToTypes 将 rpc 的 MdmRecord 逐字段转换为网关类型。
func mdmRecordToTypes(record *mdmClient.MdmRecord) *types.MdmRecordInfo {
	if record == nil {
		return &types.MdmRecordInfo{}
	}
	return &types.MdmRecordInfo{
		Id:       record.GetId(),
		ModelId:  record.GetModelId(),
		BizCode:  record.GetBizCode(),
		DataJson: record.GetDataJson(),
		Status:   record.GetStatus(),
		Version:  record.GetVersion(),
		Beid:     record.GetBeid(),
	}
}

// mdmDistributionToTypes 将 rpc 的 MdmDistribution 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func mdmDistributionToTypes(dist *mdmClient.MdmDistribution) *types.MdmDistribution {
	if dist == nil {
		return &types.MdmDistribution{}
	}
	return &types.MdmDistribution{
		Id:            dist.GetId(),
		RecordId:      dist.GetRecordId(),
		TargetDsId:    dist.GetTargetDsId(),
		TargetDb:      dist.GetTargetDb(),
		TargetTable:   dist.GetTargetTable(),
		Status:        dist.GetStatus(),
		Msg:           dist.GetMsg(),
		DistributedAt: dist.GetDistributedAt(),
	}
}
