package logic

import "errors"

// 业务错误（对齐 quality/metadata 服务写法：errors.New 中文语义错误，调用方用 errors.Is 判断）
var (
	// ErrTaskIdRequired 任务 id 缺失
	ErrTaskIdRequired = errors.New("任务 id 不能为空")
	// ErrTaskNameRequired 任务名缺失
	ErrTaskNameRequired = errors.New("任务名不能为空")
	// ErrTaskNameTooLong 任务名超长（dc_sync_task.name 为 varchar(128)）
	ErrTaskNameTooLong = errors.New("任务名长度不能超过 128 字符")
	// ErrInvalidMode 同步模式不在白名单
	ErrInvalidMode = errors.New("mode 非法：仅支持 offline/cdc")
	// ErrSourceDsRequired 源数据源缺失
	ErrSourceDsRequired = errors.New("source_ds_id 不能为空(需>0)")
	// ErrInvalidMappingJson 字段/表映射 JSON 非法
	ErrInvalidMappingJson = errors.New("mapping_json 非法：需为 JSON 对象")
	// ErrInvalidStatus 列表过滤状态不在白名单
	ErrInvalidStatus = errors.New("status 非法：仅支持 draft/deployed/running/failed")
	// ErrTaskNotFound 任务不存在（含跨租户不可见）
	ErrTaskNotFound = errors.New("同步任务不存在")
	// ErrTaskDeployed 已部署任务在调度器侧仍有工作流，需先下线才能删除
	ErrTaskDeployed = errors.New("任务已部署到调度器，请先下线后再删除")
	// ErrTaskNotDeployed 未部署任务不能手动触发
	ErrTaskNotDeployed = errors.New("任务尚未部署到调度器，请先执行部署")
)
