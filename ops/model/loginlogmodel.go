package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	loginLogTable = "`dc_ops_login_log`"
	// loginLogInsertColumns 插入字段（不含自增 id；created_at 由数据库默认当前时间）
	loginLogInsertColumns = "username,uid,ip,user_agent,success,msg"
	// loginLogRows 查询列（user_agent/msg 允许 NULL 用 ifnull 兜底；created_at 转毫秒时间戳并以 0 兜底，
	// 查询列数与 LoginLog 字段数一一对应，规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	loginLogRows = "id,username,uid,ip,ifnull(user_agent,'') as user_agent,success,ifnull(msg,'') as msg," +
		"ifnull(cast(unix_timestamp(created_at) * 1000 as signed), 0) as created_at"
)

// 登录日志（dc_ops_login_log，全平台共享，无 beid 列）
type (
	// LoginLog 登录日志行
	LoginLog struct {
		Id        int64  `db:"id"`
		Username  string `db:"username"` // 登录账号/mobile
		Uid       int64  `db:"uid"`
		Ip        string `db:"ip"`
		UserAgent string `db:"user_agent"`
		Success   int64  `db:"success"`    // 1.成功 0.失败
		Msg       string `db:"msg"`        // 失败原因等备注
		CreatedAt int64  `db:"created_at"` // 毫秒时间戳
	}

	// LoginLogModel 登录日志表操作接口
	LoginLogModel interface {
		// Insert 写入一条登录日志
		Insert(ctx context.Context, data *LoginLog) error
		// FindList 条件分页查询：username 精确匹配可选；
		// successFilter 0.全部 1.仅成功 2.仅失败；返回列表与总数
		FindList(ctx context.Context, username string, successFilter, page, size int64) ([]*LoginLog, int64, error)
	}

	defaultLoginLogModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewLoginLogModel 创建登录日志 model
func NewLoginLogModel(conn sqlx.SqlConn) LoginLogModel {
	return &defaultLoginLogModel{
		conn:  conn,
		table: loginLogTable,
	}
}

// Insert 写入一条登录日志
func (m *defaultLoginLogModel) Insert(ctx context.Context, data *LoginLog) error {
	query := fmt.Sprintf("insert into %s (%s) values (?,?,?,?,?,?)",
		m.table, loginLogInsertColumns)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Username, data.Uid, data.Ip, data.UserAgent, data.Success, data.Msg)
	return err
}

// FindList 登录日志列表（username 精确可选 + successFilter 过滤，按 id 倒序），返回列表与总数
func (m *defaultLoginLogModel) FindList(ctx context.Context, username string, successFilter, page, size int64) ([]*LoginLog, int64, error) {
	page, size = normalizePage(page, size)

	where := "1 = 1"
	args := []any{}
	if username != "" {
		where += " and `username` = ?"
		args = append(args, username)
	}
	switch successFilter {
	case 1: // 仅成功
		where += " and `success` = 1"
	case 2: // 仅失败
		where += " and `success` = 0"
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		loginLogRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*LoginLog
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
