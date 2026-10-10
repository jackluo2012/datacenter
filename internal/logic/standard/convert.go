package standard

import (
	"strconv"

	"datacenter/internal/types"
	"datacenter/standard/rpc/standardClient"
)

// wordRootToTypes 将 rpc 的 WordRoot 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func wordRootToTypes(root *standardClient.WordRoot) *types.StdWordRoot {
	if root == nil {
		return &types.StdWordRoot{}
	}
	return &types.StdWordRoot{
		Id:      root.GetId(),
		Name:    root.GetName(),
		CnName:  root.GetCnName(),
		Type:    root.GetType(),
		Comment: root.GetComment(),
	}
}

// dictItemToTypes 将 rpc 的 DictItem 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func dictItemToTypes(item *standardClient.DictItem) *types.StdDictItem {
	if item == nil {
		return &types.StdDictItem{}
	}
	return &types.StdDictItem{
		Id:        item.GetId(),
		DictId:    item.GetDictId(),
		ItemCode:  item.GetItemCode(),
		ItemValue: item.GetItemValue(),
		Sort:      item.GetSort(),
	}
}

// dictToTypes 将 rpc 的 Dict 及其 items 逐字段转换为网关类型。
// 契约差异: 网关 StdDict.Id 为 string, rpc Dict.Id 为 int64, 此处统一格式化;
// items 为 nil 时仍返回空切片, 保证 JSON 序列化为 [] 而非 null;
// dict 为 nil(items 列表场景 rpc 不回传 items)时返回零值字典与空 items。
func dictToTypes(dict *standardClient.Dict, items []*standardClient.DictItem) *types.StdDict {
	if dict == nil {
		return &types.StdDict{Items: make([]types.StdDictItem, 0)}
	}
	list := make([]types.StdDictItem, 0, len(items))
	for _, item := range items {
		list = append(list, *dictItemToTypes(item))
	}
	return &types.StdDict{
		Id:       strconv.FormatInt(dict.GetId(), 10),
		DictCode: dict.GetDictCode(),
		DictName: dict.GetDictName(),
		Comment:  dict.GetComment(),
		Status:   dict.GetStatus(),
		Items:    list,
	}
}

// elementToTypes 将 rpc 的 DataElement 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func elementToTypes(element *standardClient.DataElement) *types.StdElement {
	if element == nil {
		return &types.StdElement{}
	}
	return &types.StdElement{
		Id:            element.GetId(),
		ElementCode:   element.GetElementCode(),
		ElementName:   element.GetElementName(),
		DataType:      element.GetDataType(),
		DataLength:    element.GetDataLength(),
		DataPrecision: element.GetDataPrecision(),
		DictCode:      element.GetDictCode(),
		Comment:       element.GetComment(),
	}
}

// stdModelToTypes 将 rpc 的 StdModel 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func stdModelToTypes(model *standardClient.StdModel) *types.StdModel {
	if model == nil {
		return &types.StdModel{}
	}
	return &types.StdModel{
		Id:           model.GetId(),
		ModelCode:    model.GetModelCode(),
		ModelName:    model.GetModelName(),
		ElementsJson: model.GetElementsJson(),
		Status:       model.GetStatus(),
		Version:      model.GetVersion(),
		SnapshotJson: model.GetSnapshotJson(),
		PublishedAt:  model.GetPublishedAt(),
	}
}

// checkReportToTypes 将 rpc 的 StdCheckReport 逐字段转换为网关类型。
// rpc 侧的 Beid 为租户内部字段, 网关类型未定义, 故不做映射。
func checkReportToTypes(report *standardClient.StdCheckReport) *types.StdCheckReport {
	if report == nil {
		return &types.StdCheckReport{}
	}
	return &types.StdCheckReport{
		Id:        report.GetId(),
		ModelId:   report.GetModelId(),
		TableName: report.GetTableName(),
		DiffJson:  report.GetDiffJson(),
		PassFlag:  report.GetPassFlag(),
		RunAt:     report.GetRunAt(),
	}
}

// normalizePage 分页默认值处理: page<=0 取 1, size<=0 取 10。
func normalizePage(page, size int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return page, size
}
