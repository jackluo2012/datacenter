package logic

import (
	"unicode/utf8"

	"datacenter/ops/model"
	"datacenter/ops/rpc/ops"
)

// taskToPb 归档任务 model -> pb（batch_size 收窄 int32，delete_source/last_run_at 转换）
func taskToPb(t *model.ArchiveTask) *ops.ArchiveTask {
	return &ops.ArchiveTask{
		Id:           t.Id,
		Name:         t.Name,
		DatasourceId: t.DatasourceId,
		DbName:       t.DbName,
		TableName:    t.TableName,
		ConditionSql: t.ConditionSql,
		ArchiveTable: t.ArchiveTable,
		DeleteSource: t.DeleteSource == 1,
		BatchSize:    int32(t.BatchSize),
		Cron:         t.Cron,
		LastStatus:   t.LastStatus,
		LastRunAt:    t.LastRunAt,
		Beid:         t.Beid,
	}
}

// archiveLogToPb 归档日志 model -> pb
func archiveLogToPb(lg *model.ArchiveLog) *ops.ArchiveLog {
	return &ops.ArchiveLog{
		Id:           lg.Id,
		TaskId:       lg.TaskId,
		ArchivedRows: lg.ArchivedRows,
		DeletedRows:  lg.DeletedRows,
		Status:       lg.Status,
		Msg:          lg.Msg,
		RunAt:        lg.RunAt,
	}
}

// loginLogToPb 登录日志 model -> pb（success tinyint -> bool）
func loginLogToPb(lg *model.LoginLog) *ops.LoginLog {
	return &ops.LoginLog{
		Id:        lg.Id,
		Username:  lg.Username,
		Uid:       lg.Uid,
		Ip:        lg.Ip,
		UserAgent: lg.UserAgent,
		Success:   lg.Success == 1,
		Msg:       lg.Msg,
		CreatedAt: lg.CreatedAt,
	}
}

// auditLogToPb 审计日志 model -> pb
func auditLogToPb(a *model.AuditLog) *ops.AuditLog {
	return &ops.AuditLog{
		Id:         a.Id,
		Uid:        a.Uid,
		Username:   a.Username,
		Action:     a.Action,
		Resource:   a.Resource,
		ResourceId: a.ResourceId,
		DetailJson: a.DetailJson,
		Ip:         a.Ip,
		CreatedAt:  a.CreatedAt,
	}
}

// 日志类写接口入参的列宽上限（与 dc_ops_* 表 varchar 列宽一致，按字符计；
// 超长入参会让落库直接报错，写日志接口按列宽截断兜底，绝不能拖垮调用方主流程）
const (
	usernameMaxLen    = 128
	ipMaxLen          = 64
	userAgentMaxLen   = 512
	logMsgMaxLen      = 255
	actionMaxLen      = 64
	resourceMaxLen    = 128
	resourceIdMaxLen  = 64
	archiveMsgMaxLen  = 1024
	detailJsonMaxByte = 60000 // detail_json 为 TEXT（65535 字节），留余量按字节截断
)

// truncateRunes 按字符数截断（MySQL varchar(n) 按字符计）
func truncateRunes(s string, n int) string {
	if len(s) <= n { // 字节数都不超限，直接返回
		return s
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// truncateBytes 按字节数截断且不劈开多字节字符（TEXT 列按字节计）
func truncateBytes(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}
