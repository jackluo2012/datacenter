package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	auditLogTable = "`dc_ops_audit_log`"
	// auditLogInsertColumns 插入字段（不含自增 id；created_at 由数据库默认当前时间）
	auditLogInsertColumns = "uid,username,action,resource,resource_id,detail_json,ip"
	// auditLogRows 查询列（username/resource_id/detail_json 允许 NULL 用 ifnull 兜底；
	// created_at 转毫秒时间戳并以 0 兜底，查询列数与 AuditLog 字段数一一对应，
	// 规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	auditLogRows = "id,uid,ifnull(username,'') as username,action,resource," +
		"ifnull(resource_id,'') as resource_id,ifnull(detail_json,'') as detail_json,ip," +
		"ifnull(cast(unix_timestamp(created_at) * 1000 as signed), 0) as created_at"
)

// 通用审计日志（dc_ops_audit_log，全平台共享，无 beid 列）
type (
	// AuditLog 审计日志行
	AuditLog struct {
		Id         int64  `db:"id"`
		Uid        int64  `db:"uid"`
		Username   string `db:"username"`
		Action     string `db:"action"`   // create/update/delete/publish/login/export...
		Resource   string `db:"resource"` // 资源类型:datasource/api/rule...
		ResourceId string `db:"resource_id"`
		DetailJson string `db:"detail_json"`
		Ip         string `db:"ip"`
		CreatedAt  int64  `db:"created_at"` // 毫秒时间戳
	}

	// AuditLogModel 通用审计日志表操作接口
	AuditLogModel interface {
		// Insert 写入一条审计日志
		Insert(ctx context.Context, data *AuditLog) error
		// FindList 条件分页查询：resource 精确匹配可选、uid>0 过滤可选；返回列表与总数
		FindList(ctx context.Context, resource string, uid, page, size int64) ([]*AuditLog, int64, error)
	}

	defaultAuditLogModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewAuditLogModel 创建通用审计日志 model
func NewAuditLogModel(conn sqlx.SqlConn) AuditLogModel {
	return &defaultAuditLogModel{
		conn:  conn,
		table: auditLogTable,
	}
}

// Insert 写入一条审计日志
func (m *defaultAuditLogModel) Insert(ctx context.Context, data *AuditLog) error {
	query := fmt.Sprintf("insert into %s (%s) values (?,?,?,?,?,?,?)",
		m.table, auditLogInsertColumns)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Uid, data.Username, data.Action, data.Resource,
		data.ResourceId, data.DetailJson, data.Ip)
	return err
}

// FindList 审计日志列表（resource/uid 可选过滤，按 id 倒序），返回列表与总数
func (m *defaultAuditLogModel) FindList(ctx context.Context, resource string, uid, page, size int64) ([]*AuditLog, int64, error) {
	page, size = normalizePage(page, size)

	where := "1 = 1"
	args := []any{}
	if resource != "" {
		where += " and `resource` = ?"
		args = append(args, resource)
	}
	if uid > 0 {
		where += " and `uid` = ?"
		args = append(args, uid)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		auditLogRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*AuditLog
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
