package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	checkReportTable           = "`dc_std_check_report`"
	checkReportInsertFieldSize = 8 // 插入字段数（与 checkReportInsertColumns 对应）
	// checkReportInsertColumns 插入字段（不含自增 id，run_at 显式传入保证落库时间一致）
	checkReportInsertColumns = "`model_id`,`datasource_id`,`db_name`,`table_name`,`diff_json`,`pass_flag`,`run_at`,`beid`"
	// checkReportRows 查询列（9 列与 CheckReport 9 字段一一对应，
	// 规避 go-zero 少列查询扫宽 struct 报 not matching destination to scan；
	// db_name/diff_json 允许 NULL 用 ifnull 兜底，run_at 转毫秒时间戳并以 0 兜底）
	checkReportRows = "`id`,`model_id`,`datasource_id`,ifnull(`db_name`,'') as `db_name`,`table_name`," +
		"ifnull(`diff_json`,'') as `diff_json`,`pass_flag`," +
		"ifnull(cast(unix_timestamp(`run_at`) * 1000 as signed), 0) as `run_at`,`beid`"
)

// 标准核对报告（dc_std_check_report）
type (
	// CheckReport 核对报告行（不映射 created_at，查询列与字段数保持 9/9 严格对齐）
	CheckReport struct {
		Id           int64  `db:"id"`
		ModelId      int64  `db:"model_id"`
		DatasourceId int64  `db:"datasource_id"`
		DbName       string `db:"db_name"`
		TableName    string `db:"table_name"`
		DiffJson     string `db:"diff_json"` // 差异 [{column,expect,actual,result}]
		PassFlag     int64  `db:"pass_flag"` // 1.符合 0.存在偏差
		RunAt        int64  `db:"run_at"`    // 执行时间（毫秒时间戳）
		Beid         int64  `db:"beid"`
	}

	// CheckReportModel 核对报告表操作接口
	CheckReportModel interface {
		// Insert 新增核对报告，返回自增 id
		Insert(ctx context.Context, data *CheckReport) (int64, error)
		// FindListByModel 按标准模型分页查询核对历史，返回列表与总数
		FindListByModel(ctx context.Context, beid, modelId, page, size int64) ([]*CheckReport, int64, error)
	}

	defaultCheckReportModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewCheckReportModel 创建核对报告 model
func NewCheckReportModel(conn sqlx.SqlConn) CheckReportModel {
	return &defaultCheckReportModel{
		conn:  conn,
		table: checkReportTable,
	}
}

// Insert 新增核对报告（run_at 结构体存毫秒时间戳，落库转 datetime）
func (m *defaultCheckReportModel) Insert(ctx context.Context, data *CheckReport) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, checkReportInsertColumns, placeholders(checkReportInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query,
		data.ModelId, data.DatasourceId, data.DbName, data.TableName,
		data.DiffJson, data.PassFlag, time.UnixMilli(data.RunAt), data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindListByModel 按标准模型分页查询核对历史（beid 多租户过滤），返回列表与总数
func (m *defaultCheckReportModel) FindListByModel(ctx context.Context, beid, modelId,
	page, size int64) ([]*CheckReport, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	if modelId <= 0 {
		return nil, 0, ErrModelIdRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ? and `model_id` = ?"
	args := []any{beid, modelId}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		checkReportRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*CheckReport
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
