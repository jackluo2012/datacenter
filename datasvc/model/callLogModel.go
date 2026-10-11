package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 调用结果状态（dc_api_call_log.status 取值）
const (
	CallStatusOk     = "ok"
	CallStatusError  = "error"
	CallStatusDenied = "denied"
)

// callLogInsertCols 调用日志写入列（created_at 由 DB DEFAULT CURRENT_TIMESTAMP(3) 生成）
const callLogInsertCols = "`api_id`, `app_id`, `path`, `status`, `error_code`, `rows_cnt`, `cost_ms`, `cache_hit`, `ip`"

// callLogSelectRows 调用日志查询列（path/error_code/ip 有 default ” 兜底，无 NULL 风险）
const callLogSelectRows = "l.`id`, l.`api_id`, l.`app_id`, l.`path`, l.`status`, " +
	"l.`error_code`, l.`rows_cnt`, l.`cost_ms`, l.`cache_hit`, l.`ip`, l.`created_at`"

type (
	// CallLogModel dc_api_call_log 表模型（API 可观测：调用日志 + 统计聚合）
	CallLogModel interface {
		// Insert 落一条调用日志（ExecuteApi 埋点单条写入），返回自增 id
		Insert(data *ApiCallLog) (int64, error)
		// BatchInsert 批量落调用日志（离线补录/数据修复用），拼一次多值 insert
		BatchInsert(list []*ApiCallLog) error
		// FindList 调用日志分页查询：beid/api_id/app_id/status 多维可选过滤，时间倒序
		FindList(beid, apiId, appId int64, status string, page, size int64) ([]*ApiCallLog, int64, error)
		// Stats 窗口 [begin,end) 内一次 SQL 聚合：总调用/ok/error/denied 计数、平均耗时、cache 命中数
		Stats(beid int64, begin, end time.Time) (*ApiCallStats, error)
		// TopApis 窗口内调用量 TopN（join dc_api 出 path），按调用量倒序
		TopApis(beid int64, begin, end time.Time, limit int64) ([]*ApiTopCall, error)
		// DailyCalls 窗口内按天调用量（按 date(created_at) 归组），升序
		DailyCalls(beid int64, begin, end time.Time) ([]*ApiDailyCall, error)
	}

	defaultCallLogModel struct {
		conn  sqlx.SqlConn
		table string
	}

	// ApiCallLog dc_api_call_log 表结构
	ApiCallLog struct {
		Id        int64     `db:"id"`
		ApiId     int64     `db:"api_id"`
		AppId     int64     `db:"app_id"`
		Path      string    `db:"path"`
		Status    string    `db:"status"` //ok/error/denied
		ErrorCode string    `db:"error_code"`
		RowsCnt   int64     `db:"rows_cnt"` //返回行数/写语句受影响行数
		CostMs    int64     `db:"cost_ms"`
		CacheHit  int64     `db:"cache_hit"` //0/1
		Ip        string    `db:"ip"`
		CreatedAt time.Time `db:"created_at"`
	}

	// ApiCallStats 调用统计聚合行（一次 SQL 汇出，供 GetCallStats 总览）
	ApiCallStats struct {
		TotalCalls  int64   `db:"total_calls"`
		OkCalls     int64   `db:"ok_calls"`
		ErrorCalls  int64   `db:"error_calls"`
		DeniedCalls int64   `db:"denied_calls"`
		AvgCostMs   float64 `db:"avg_cost_ms"`
		CacheHits   int64   `db:"cache_hits"`
	}

	// ApiTopCall API 调用量排行行
	ApiTopCall struct {
		ApiId     int64   `db:"api_id"`
		Path      string  `db:"path"` //join dc_api 所得，API 已删除时为空串
		Calls     int64   `db:"calls"`
		AvgCostMs float64 `db:"avg_cost_ms"`
	}

	// ApiDailyCall 按天调用量行（day 格式 YYYY-MM-DD）
	ApiDailyCall struct {
		Day   string `db:"day"`
		Calls int64  `db:"calls"`
	}
)

func NewCallLogModel(conn sqlx.SqlConn) CallLogModel {
	return &defaultCallLogModel{
		conn:  conn,
		table: "`dc_api_call_log`",
	}
}

// Insert 落一条调用日志，返回自增 id
func (m *defaultCallLogModel) Insert(data *ApiCallLog) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		m.table, callLogInsertCols)
	ret, err := m.conn.Exec(query,
		data.ApiId, data.AppId, data.Path, data.Status, data.ErrorCode,
		data.RowsCnt, data.CostMs, data.CacheHit, data.Ip)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// BatchInsert 批量落调用日志：单条多值 insert，减少高峰期写放大
func (m *defaultCallLogModel) BatchInsert(list []*ApiCallLog) error {
	if len(list) == 0 {
		return nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "insert into %s (%s) values ", m.table, callLogInsertCols)
	args := make([]interface{}, 0, len(list)*9)
	for i, d := range list {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args, d.ApiId, d.AppId, d.Path, d.Status, d.ErrorCode,
			d.RowsCnt, d.CostMs, d.CacheHit, d.Ip)
	}
	_, err := m.conn.Exec(sb.String(), args...)
	return err
}

// beidCond 租户过滤条件：dc_api_call_log 无 beid 列，经 exists 子查询关联 dc_api 收敛；
// beid=0 表示不过滤（跨租户全量，供运维视角）
func beidCond(beid int64) (string, []interface{}) {
	if beid == 0 {
		return "", nil
	}
	return " and exists(select 1 from `dc_api` api_scope " +
		"where api_scope.`id` = l.`api_id` and api_scope.`beid` = ?)", []interface{}{beid}
}

// FindList 调用日志分页查询：beid/api_id/app_id/status 全部可选（零值不过滤），时间倒序
func (m *defaultCallLogModel) FindList(beid, apiId, appId int64, status string, page, size int64) ([]*ApiCallLog, int64, error) {
	//分页参数兜底（对齐 apiModel.List）
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	where := "1=1"
	var args []interface{}
	cond, cargs := beidCond(beid)
	where += cond
	args = append(args, cargs...)
	if apiId > 0 {
		where += " and l.`api_id` = ?"
		args = append(args, apiId)
	}
	if appId > 0 {
		where += " and l.`app_id` = ?"
		args = append(args, appId)
	}
	if status = strings.TrimSpace(status); status != "" {
		where += " and l.`status` = ?"
		args = append(args, status)
	}
	var total int64
	if err := m.conn.QueryRow(&total,
		fmt.Sprintf("select count(*) from %s l where %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*ApiCallLog{}, 0, nil
	}
	query := fmt.Sprintf(
		"select %s from %s l where %s order by l.`created_at` desc, l.`id` desc limit ? offset ?",
		callLogSelectRows, m.table, where)
	args = append(args, size, (page-1)*size)
	var list []*ApiCallLog
	if err := m.conn.QueryRows(&list, query, args...); err != nil {
		switch err {
		case sqlc.ErrNotFound:
			return []*ApiCallLog{}, total, nil
		default:
			return nil, 0, err
		}
	}
	return list, total, nil
}

// Stats 一次 SQL 聚合统计（窗口 [begin,end)）：sum/ifnull 兜底空窗口返回 0 而非 NULL
func (m *defaultCallLogModel) Stats(beid int64, begin, end time.Time) (*ApiCallStats, error) {
	cond, cargs := beidCond(beid)
	query := fmt.Sprintf(
		"select count(*) as `total_calls`, "+
			"ifnull(sum(case when l.`status` = 'ok' then 1 else 0 end), 0) as `ok_calls`, "+
			"ifnull(sum(case when l.`status` = 'error' then 1 else 0 end), 0) as `error_calls`, "+
			"ifnull(sum(case when l.`status` = 'denied' then 1 else 0 end), 0) as `denied_calls`, "+
			"ifnull(avg(l.`cost_ms`), 0) as `avg_cost_ms`, "+
			"ifnull(sum(l.`cache_hit`), 0) as `cache_hits` "+
			"from %s l where l.`created_at` >= ? and l.`created_at` < ?%s",
		m.table, cond)
	args := append([]interface{}{begin, end}, cargs...)
	var resp ApiCallStats
	if err := m.conn.QueryRow(&resp, query, args...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TopApis 调用量 TopN：left join dc_api 补 path（API 已删除时 path 置空串，不丢调用记录）
func (m *defaultCallLogModel) TopApis(beid int64, begin, end time.Time, limit int64) ([]*ApiTopCall, error) {
	cond, cargs := beidCond(beid)
	query := fmt.Sprintf(
		"select l.`api_id` as `api_id`, ifnull(a.`path`, '') as `path`, "+
			"count(*) as `calls`, ifnull(avg(l.`cost_ms`), 0) as `avg_cost_ms` "+
			"from %s l left join `dc_api` a on a.`id` = l.`api_id` "+
			"where l.`created_at` >= ? and l.`created_at` < ?%s "+
			"group by l.`api_id`, a.`path` order by `calls` desc, l.`api_id` asc limit ?",
		m.table, cond)
	args := append([]interface{}{begin, end}, cargs...)
	args = append(args, limit)
	var list []*ApiTopCall
	err := m.conn.QueryRows(&list, query, args...)
	switch err {
	case nil:
		return list, nil
	case sqlc.ErrNotFound:
		return []*ApiTopCall{}, nil
	default:
		return nil, err
	}
}

// DailyCalls 按天调用量：select/group by 均用 date_format 输出 YYYY-MM-DD 字符串——
// 语义等价按 date(created_at) 归组，同时规避 DSN parseTime=true 下 DATE 列扫描为 time.Time
func (m *defaultCallLogModel) DailyCalls(beid int64, begin, end time.Time) ([]*ApiDailyCall, error) {
	cond, cargs := beidCond(beid)
	query := fmt.Sprintf(
		"select date_format(l.`created_at`, '%%Y-%%m-%%d') as `day`, count(*) as `calls` "+
			"from %s l where l.`created_at` >= ? and l.`created_at` < ?%s "+
			"group by date_format(l.`created_at`, '%%Y-%%m-%%d') order by `day` asc",
		m.table, cond)
	args := append([]interface{}{begin, end}, cargs...)
	var list []*ApiDailyCall
	err := m.conn.QueryRows(&list, query, args...)
	switch err {
	case nil:
		return list, nil
	case sqlc.ErrNotFound:
		return []*ApiDailyCall{}, nil
	default:
		return nil, err
	}
}
