package logic

import (
	"datacenter/mdm/model"
	"datacenter/mdm/rpc/mdm"
)

// toPbModel model → pb（主数据模型）
func toPbModel(m *model.MdmModelRow) *mdm.MdmModel {
	if m == nil {
		return nil
	}
	return &mdm.MdmModel{
		Id:           m.Id,
		ModelCode:    m.ModelCode,
		ModelName:    m.ModelName,
		TableName:    m.TableName,
		FieldsJson:   m.FieldsJson,
		CodeStrategy: m.CodeStrategy,
		Beid:         m.Beid,
	}
}

// toPbRecord model → pb（主数据记录）
func toPbRecord(r *model.RecordRow) *mdm.MdmRecord {
	if r == nil {
		return nil
	}
	return &mdm.MdmRecord{
		Id:       r.Id,
		ModelId:  r.ModelId,
		BizCode:  r.BizCode,
		DataJson: r.DataJson,
		Status:   r.Status,
		Version:  r.Version,
		Beid:     r.Beid,
	}
}

// toPbDistribution model → pb（分发记录）
func toPbDistribution(d *model.DistributionRow) *mdm.MdmDistribution {
	if d == nil {
		return nil
	}
	return &mdm.MdmDistribution{
		Id:            d.Id,
		RecordId:      d.RecordId,
		TargetDsId:    d.TargetDsId,
		TargetDb:      d.TargetDb,
		TargetTable:   d.TargetTable,
		Status:        d.Status,
		Msg:           d.Msg,
		DistributedAt: d.DistributedAt,
		Beid:          d.Beid,
	}
}
