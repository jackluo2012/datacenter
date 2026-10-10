package logic

import (
	"datacenter/standard/model"
	"datacenter/standard/rpc/standard"
)

// rootToPb 词根 model -> pb
func rootToPb(r *model.WordRoot) *standard.WordRoot {
	return &standard.WordRoot{
		Id:      r.Id,
		Name:    r.Name,
		CnName:  r.CnName,
		Type:    r.Type,
		Comment: r.Comment,
		Beid:    r.Beid,
	}
}

// dictToPb 字典 model -> pb（status 为 int32 收窄转换）
func dictToPb(d *model.Dict) *standard.Dict {
	return &standard.Dict{
		Id:       d.Id,
		DictCode: d.DictCode,
		DictName: d.DictName,
		Comment:  d.Comment,
		Status:   int32(d.Status),
		Beid:     d.Beid,
	}
}

// itemToPb 字典项 model -> pb（sort 为 int32 收窄转换）
func itemToPb(i *model.DictItem) *standard.DictItem {
	return &standard.DictItem{
		Id:        i.Id,
		DictId:    i.DictId,
		ItemCode:  i.ItemCode,
		ItemValue: i.ItemValue,
		Sort:      int32(i.Sort),
		Beid:      i.Beid,
	}
}

// elementToPb 数据元 model -> pb
func elementToPb(e *model.Element) *standard.DataElement {
	return &standard.DataElement{
		Id:            e.Id,
		ElementCode:   e.ElementCode,
		ElementName:   e.ElementName,
		DataType:      e.DataType,
		DataLength:    e.DataLength,
		DataPrecision: e.DataPrecision,
		DictCode:      e.DictCode,
		Comment:       e.Comment,
		Beid:          e.Beid,
	}
}

// modelToPb 标准模型 model -> pb（version 为 int32 收窄转换）
func modelToPb(m *model.StdModel) *standard.StdModel {
	return &standard.StdModel{
		Id:           m.Id,
		ModelCode:    m.ModelCode,
		ModelName:    m.ModelName,
		ElementsJson: m.ElementsJson,
		Status:       m.Status,
		Version:      m.Version,
		SnapshotJson: m.SnapshotJson,
		PublishedAt:  m.PublishedAt,
		Beid:         m.Beid,
	}
}

// reportToPb 核对报告 model -> pb（pass_flag tinyint -> bool）
func reportToPb(r *model.CheckReport) *standard.StdCheckReport {
	return &standard.StdCheckReport{
		Id:        r.Id,
		ModelId:   r.ModelId,
		TableName: r.TableName,
		DiffJson:  r.DiffJson,
		PassFlag:  r.PassFlag == 1,
		RunAt:     r.RunAt,
	}
}
