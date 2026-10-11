package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	profileTable           = "`dc_quality_profile`"
	profileInsertFieldSize = 13 // 插入字段数（与 profileInsertColumns 对应）
	// profileInsertColumns 插入字段（不含自增 id，ran_at 显式传入保证与剖析执行时间一致）
	profileInsertColumns = "datasource_id,db_name,table_name,rows_cnt,cols_cnt,null_cells,total_cells," +
		"null_rate,dup_pk_cnt,health_score,profile_json,ran_at,beid"
	// profileRows 查询列（profile_json 允许 NULL 用 ifnull 兜底；ran_at 转毫秒时间戳并以 0 兜底；
	// 查询列数与 Profile 字段数一一对应，规避 go-zero 少列扫宽 struct 的 not matching destination 报错）
	profileRows = "id,datasource_id,db_name,table_name,rows_cnt,cols_cnt,null_cells,total_cells," +
		"null_rate,dup_pk_cnt,health_score,ifnull(profile_json,'') as profile_json," +
		"ifnull(cast(unix_timestamp(ran_at) * 1000 as signed), 0) as ran_at"
)

// 表剖析与健康分（dc_quality_profile）
type (
	// Profile 表剖析结果行
	Profile struct {
		Id           int64   `db:"id"`
		DatasourceId int64   `db:"datasource_id"`
		DbName       string  `db:"db_name"`
		TableName    string  `db:"table_name"`
		RowsCnt      int64   `db:"rows_cnt"`    // 总行数
		ColsCnt      int64   `db:"cols_cnt"`    // 总列数
		NullCells    int64   `db:"null_cells"`  // 空值单元格数
		TotalCells   int64   `db:"total_cells"` // 总单元格数
		NullRate     float64 `db:"null_rate"`   // 空值率 0-1
		DupPkCnt     int64   `db:"dup_pk_cnt"`  // 主键重复组数（未传 pk_column 为 0）
		HealthScore  int64   `db:"health_score"`
		ProfileJson  string  `db:"profile_json"` // 逐列剖析明细
		RanAt        int64   `db:"ran_at"`       // 剖析时间（毫秒时间戳）
		Beid         int64   `db:"beid"`
	}

	// ProfileModel 表剖析与健康分表操作接口
	ProfileModel interface {
		// Insert 写入一次剖析结果，返回自增 id
		Insert(ctx context.Context, data *Profile) (int64, error)
		// FindList 剖析记录分页：beid 必带，datasourceId/tableName 可选过滤；返回列表与总数
		FindList(ctx context.Context, beid, datasourceId, page, size int64, tableName string) ([]*Profile, int64, error)
	}

	defaultProfileModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewProfileModel 创建表剖析 model
func NewProfileModel(conn sqlx.SqlConn) ProfileModel {
	return &defaultProfileModel{
		conn:  conn,
		table: profileTable,
	}
}

// Insert 写入一次剖析结果
func (m *defaultProfileModel) Insert(ctx context.Context, data *Profile) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, profileInsertColumns, placeholders(profileInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query,
		data.DatasourceId, data.DbName, data.TableName,
		data.RowsCnt, data.ColsCnt, data.NullCells, data.TotalCells,
		data.NullRate, data.DupPkCnt, data.HealthScore, data.ProfileJson,
		time.UnixMilli(data.RanAt), data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindList 剖析记录分页（beid 多租户必带 + datasource_id/table_name 可选过滤），返回列表与总数
func (m *defaultProfileModel) FindList(ctx context.Context, beid, datasourceId, page, size int64,
	tableName string) ([]*Profile, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if datasourceId > 0 {
		where += " and `datasource_id` = ?"
		args = append(args, datasourceId)
	}
	if tableName != "" {
		where += " and `table_name` = ?"
		args = append(args, tableName)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		profileRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Profile
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
