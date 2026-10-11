package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 告警发送结果（dc_quality_alert_log.status）
const (
	AlertSendSuccess = "success" // 机器人接收成功
	AlertSendFailed  = "failed"  // 发送失败（网络/机器人报错）
)

const (
	alertLogTable           = "`dc_quality_alert_log`"
	alertLogInsertFieldSize = 6 // 插入字段数（与 alertLogInsertColumns 对应）
	// alertLogInsertColumns 插入字段（不含自增 id，sent_at 显式传入保证与任务执行时间对齐）
	alertLogInsertColumns = "channel_id,job_id,content,status,msg,sent_at"
)

// 告警发送日志（dc_quality_alert_log）
type (
	// AlertLog 告警发送日志行
	AlertLog struct {
		Id        int64  `db:"id"`
		ChannelId int64  `db:"channel_id"`
		JobId     int64  `db:"job_id"`
		Content   string `db:"content"` // 发送的告警文本（截断后）
		Status    string `db:"status"`  // success/failed
		Msg       string `db:"msg"`     // 失败原因（成功为空）
		SentAt    int64  `db:"sent_at"` // 发送时间（毫秒时间戳）
	}

	// AlertLogModel 告警发送日志表操作接口
	AlertLogModel interface {
		// Insert 记录一次发送结果（发送失败也落库，只记日志不抛错）
		Insert(ctx context.Context, data *AlertLog) error
	}

	defaultAlertLogModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewAlertLogModel 创建告警发送日志 model
func NewAlertLogModel(conn sqlx.SqlConn) AlertLogModel {
	return &defaultAlertLogModel{
		conn:  conn,
		table: alertLogTable,
	}
}

// Insert 记录一次告警发送结果
func (m *defaultAlertLogModel) Insert(ctx context.Context, data *AlertLog) error {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, alertLogInsertColumns, placeholders(alertLogInsertFieldSize))
	_, err := m.conn.ExecCtx(ctx, query,
		data.ChannelId, data.JobId, data.Content, data.Status, data.Msg, time.UnixMilli(data.SentAt))
	return err
}
