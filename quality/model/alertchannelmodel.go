package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 告警通道类型（dc_quality_alert_channel.type 白名单，email 预留未开放）
const (
	ChannelTypeDingtalk = "dingtalk" // 钉钉群机器人
	ChannelTypeFeishu   = "feishu"   // 飞书群机器人
)

// 告警触发级别（dc_quality_alert_channel.min_level）
const (
	AlertLevelFailed  = "failed"  // 仅任务全部不通过才触发
	AlertLevelPartial = "partial" // 部分不通过即触发（含 failed）
)

// 通道启用状态（dc_quality_alert_channel.status）
const (
	ChannelStatusEnabled  = 1 // 启用
	ChannelStatusDisabled = 0 // 停用
)

// IsValidChannelType 通道类型是否在白名单内
func IsValidChannelType(t string) bool {
	switch t {
	case ChannelTypeDingtalk, ChannelTypeFeishu:
		return true
	default:
		return false
	}
}

// IsValidAlertLevel 触发级别是否在白名单内
func IsValidAlertLevel(level string) bool {
	switch level {
	case AlertLevelFailed, AlertLevelPartial:
		return true
	default:
		return false
	}
}

const (
	channelTable           = "`dc_quality_alert_channel`"
	channelInsertFieldSize = 7 // 插入字段数（与 channelInsertColumns 对应）
	// channelInsertColumns 插入字段（不含自增 id 与时间列），与 Insert 传参顺序一一对应
	channelInsertColumns = "name,type,webhook_url,secret,min_level,status,beid"
	// channelRows 查询列（secret/min_level 列定义带默认值仍用 ifnull 兜底避免 NULL 扫描报错；
	// 查询列数与 AlertChannel 字段数一一对应，规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	channelRows = "id,name,type,webhook_url,ifnull(secret,'') as secret,ifnull(min_level,'') as min_level," +
		"status,beid,created_at,updated_at"
)

// 告警通道（dc_quality_alert_channel）
type (
	// AlertChannel 告警通道行
	AlertChannel struct {
		Id         int64     `db:"id"`
		Name       string    `db:"name"`
		Type       string    `db:"type"`        // dingtalk/feishu
		WebhookUrl string    `db:"webhook_url"` // 机器人 webhook 地址
		Secret     string    `db:"secret"`      // 加签密钥（可选）
		MinLevel   string    `db:"min_level"`   // 触发级别 failed/partial
		Status     int64     `db:"status"`      // 1.启用 0.停用
		Beid       int64     `db:"beid"`
		CreatedAt  time.Time `db:"created_at"`
		UpdatedAt  time.Time `db:"updated_at"`
	}

	// AlertChannelModel 告警通道表操作接口
	AlertChannelModel interface {
		// Insert 新增通道，返回自增 id
		Insert(ctx context.Context, data *AlertChannel) (int64, error)
		// Update 更新通道可编辑字段（按 id+beid 定位，不改启用状态）
		Update(ctx context.Context, data *AlertChannel) error
		// Delete 删除通道（按 id+beid 定位）
		Delete(ctx context.Context, id, beid int64) error
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*AlertChannel, error)
		// FindEnabled 查询租户下全部启用通道（RunJob 告警分发用）
		FindEnabled(ctx context.Context, beid int64) ([]*AlertChannel, error)
		// FindList 通道分页列表：beid 必带；返回列表与总数
		FindList(ctx context.Context, beid, page, size int64) ([]*AlertChannel, int64, error)
	}

	defaultAlertChannelModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewAlertChannelModel 创建告警通道 model
func NewAlertChannelModel(conn sqlx.SqlConn) AlertChannelModel {
	return &defaultAlertChannelModel{
		conn:  conn,
		table: channelTable,
	}
}

// Insert 新增告警通道
func (m *defaultAlertChannelModel) Insert(ctx context.Context, data *AlertChannel) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, channelInsertColumns, placeholders(channelInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query,
		strings.TrimSpace(data.Name), data.Type, strings.TrimSpace(data.WebhookUrl),
		data.Secret, data.MinLevel, data.Status, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// Update 更新通道可编辑字段（启用状态不在保存接口里变更，避免误停用）
func (m *defaultAlertChannelModel) Update(ctx context.Context, data *AlertChannel) error {
	query := fmt.Sprintf("update %s set `name` = ?, `type` = ?, `webhook_url` = ?, "+
		"`secret` = ?, `min_level` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		strings.TrimSpace(data.Name), data.Type, strings.TrimSpace(data.WebhookUrl),
		data.Secret, data.MinLevel, data.Id, data.Beid)
	return err
}

// Delete 删除通道（历史告警日志不级联清理，供发送记录追溯）
func (m *defaultAlertChannelModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindOne 按主键查询通道
func (m *defaultAlertChannelModel) FindOne(ctx context.Context, id, beid int64) (*AlertChannel, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", channelRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp AlertChannel
	err := m.conn.QueryRowCtx(ctx, &resp, query+" limit 1", args...)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindEnabled 查询租户下全部启用通道（按 id 升序保证分发顺序稳定）
func (m *defaultAlertChannelModel) FindEnabled(ctx context.Context, beid int64) ([]*AlertChannel, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `status` = ? order by `id` asc",
		channelRows, m.table)
	var list []*AlertChannel
	if err := m.conn.QueryRowsCtx(ctx, &list, query, beid, ChannelStatusEnabled); err != nil {
		return nil, err
	}
	return list, nil
}

// FindList 通道分页列表（beid 多租户必带），返回列表与总数
func (m *defaultAlertChannelModel) FindList(ctx context.Context, beid, page, size int64) ([]*AlertChannel, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where `beid` = ?", m.table)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, beid); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where `beid` = ? order by `id` desc limit ?, ?",
		channelRows, m.table)
	var list []*AlertChannel
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, beid, (page-1)*size, size); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
