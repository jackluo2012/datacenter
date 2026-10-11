package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/datasource/rpc/datasource"
	"datacenter/quality/model"
	"datacenter/quality/rpc/quality"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 健康分扣分权重与自定义扣分项（全部为正数扣分）
const (
	// healthNullWeight 空值率权重：空值率 1.0（全空）扣 40 分
	healthNullWeight = 40.0
	// healthDupWeight 主键重复占比权重：重复占比 1.0 扣 30 分
	healthDupWeight = 30.0
	// emptyTableDeduct 空表额外扣分（自定义扣分项：0 行表对数仓无分析价值）
	emptyTableDeduct = 10.0
)

// ErrProfileDbUnknown 无法确定剖析目标库（请求未带 db_name 且 DSN 中也不含库名）
var ErrProfileDbUnknown = fmt.Errorf("%w: 请传 db_name 或使用含库名的数据源 DSN", ErrDataSourceDsn)

// errInvalidIdentifier 构造字段级标识符白名单错误（字段名 + 原值便于定位）
func errInvalidIdentifier(field, value string) error {
	return fmt.Errorf("%w: %s %q", ErrInvalidIdentifier, field, value)
}

// profileTargetConn 从数据源服务获取 DSN 并建立目标库连接（与 RunJob 同装配模式），
// 返回连接与原始 DSN（供库名解析）
func profileTargetConn(ctx context.Context, cli dataSourceClient.DataSource,
	datasourceId, beid int64) (sqlx.SqlConn, string, error) {
	reply, err := cli.GetDataSource(ctx, &datasource.GetDataSourceReq{
		Id:         datasourceId,
		Beid:       beid,
		WithSecret: true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("获取数据源失败: %w", err)
	}
	if reply == nil || reply.Datasource == nil || reply.Dsn == "" {
		return nil, "", ErrDataSourceDsn
	}
	if reply.Datasource.Type != "mysql" {
		return nil, "", fmt.Errorf("%w: 表剖析仅支持 mysql 数据源，当前 %q", ErrDataSourceDsn, reply.Datasource.Type)
	}
	return sqlx.NewMysql(reply.Dsn), reply.Dsn, nil
}

// ProfileDbName 解析剖析目标库名（纯函数）：请求显式指定优先，否则从 DSN 解析默认库；
// 两者都拿不到返回 ErrProfileDbUnknown
func ProfileDbName(reqDb, dsn string) (string, error) {
	if db := strings.TrimSpace(reqDb); db != "" {
		return db, nil
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil || strings.TrimSpace(cfg.DBName) == "" {
		return "", ErrProfileDbUnknown
	}
	return cfg.DBName, nil
}

// HealthScore 健康分计算（纯函数）：
//
//	score = 100 - 空值率×40 - 主键重复占比×30 - 自定义扣分项，四舍五入后夹取 [0,100]
//
// 入参均为 0-1 的比例值（extraDeduct 为任意正数扣分项），负数入参按 0 处理
func HealthScore(nullRate, dupPkRatio, extraDeduct float64) int32 {
	if nullRate < 0 {
		nullRate = 0
	}
	if dupPkRatio < 0 {
		dupPkRatio = 0
	}
	if extraDeduct < 0 {
		extraDeduct = 0
	}
	score := 100.0 - nullRate*healthNullWeight - dupPkRatio*healthDupWeight - extraDeduct
	return int32(math.Round(math.Min(100, math.Max(0, score))))
}

// ratio 比例计算（纯函数）：total<=0 返回 0（空表/空单元格不产生扣分，空表另走自定义扣分项）
func ratio(part, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(part) / float64(total)
}

// profileColumn information_schema 剖析出的目标表列
type profileColumn struct {
	Column   string `db:"column_name"`
	DataType string `db:"data_type"`
}

// columnProfile 单列剖析明细（profile_json 数组元素）
type columnProfile struct {
	Column      string  `json:"column"`
	NullCnt     int64   `json:"null_cnt"`
	NullRate    float64 `json:"null_rate"`
	DistinctCnt int64   `json:"distinct_cnt"`
	SampleMin   string  `json:"sample_min"`
	SampleMax   string  `json:"sample_max"`
}

// colAgg 单列聚合结果（扫描载体，与 listColumnProfile 的查询列一一对应）
type colAgg struct {
	NullCnt     int64  `db:"null_cnt"`
	DistinctCnt int64  `db:"distinct_cnt"`
	SampleMin   string `db:"sample_min"`
	SampleMax   string `db:"sample_max"`
}

// BuildColumnProfileJSON 逐列明细序列化为 profile_json（纯函数，序列化失败降级为空串）
func BuildColumnProfileJSON(cols []*columnProfile) string {
	if cols == nil {
		cols = []*columnProfile{}
	}
	b, err := json.Marshal(cols)
	if err != nil {
		return ""
	}
	return string(b)
}

// profileToPb 剖析 model -> pb（cols_cnt/health_score 为 int32 收窄转换）
func profileToPb(p *model.Profile) *quality.TableProfile {
	return &quality.TableProfile{
		Id:           p.Id,
		DatasourceId: p.DatasourceId,
		TableName:    p.TableName,
		RowsCnt:      p.RowsCnt,
		ColsCnt:      int32(p.ColsCnt),
		NullRate:     p.NullRate,
		DupPkCnt:     p.DupPkCnt,
		HealthScore:  int32(p.HealthScore),
		ProfileJson:  p.ProfileJson,
		RanAt:        p.RanAt,
	}
}

// listTableColumns 查询目标表的列清单（information_schema，库名/表名走占位符防注入）
func listTableColumns(ctx context.Context, conn sqlx.SqlConn, db, table string) ([]*profileColumn, error) {
	var cols []*profileColumn
	query := "select column_name, data_type from information_schema.columns " +
		"where table_schema = ? and table_name = ? order by ordinal_position"
	if err := conn.QueryRowsCtx(ctx, &cols, query, db, table); err != nil {
		return nil, fmt.Errorf("查询表结构失败: %w", err)
	}
	if len(cols) == 0 {
		return nil, model.ErrNotFound
	}
	return cols, nil
}

// tableExists 目标表是否存在（information_schema 计数）
func tableExists(ctx context.Context, conn sqlx.SqlConn, db, table string) (bool, error) {
	var cnt int64
	query := "select count(*) from information_schema.tables where table_schema = ? and table_name = ?"
	if err := conn.QueryRowCtx(ctx, &cnt, query, db, table); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// profileColumnAgg 剖析单列：空值数（null 或空串）+ 去重数 + 最小/最大样本。
// 列名已过标识符白名单后拼入 SQL，聚合结果列固定别名与 colAgg 字段对齐
func profileColumnAgg(ctx context.Context, conn sqlx.SqlConn, target, column string) (*colAgg, error) {
	var agg colAgg
	query := fmt.Sprintf("select cast(ifnull(sum(`%s` is null or `%s` = ''), 0) as signed) as null_cnt, "+
		"count(distinct `%s`) as distinct_cnt, "+
		"ifnull(min(`%s`), '') as sample_min, ifnull(max(`%s`), '') as sample_max "+
		"from %s",
		column, column, column, column, column, target)
	if err := conn.QueryRowCtx(ctx, &agg, query); err != nil {
		return nil, fmt.Errorf("列 %s 剖析失败: %w", column, err)
	}
	return &agg, nil
}

// countDuplicateGroups 主键重复组数（与 unique 规则同口径：group by having count(*)>1 计组数）
func countDuplicateGroups(ctx context.Context, conn sqlx.SqlConn, target, pkColumn string) (int64, error) {
	var cnt int64
	query := fmt.Sprintf("select count(*) from (select `%s` from %s group by `%s` having count(*) > 1) t",
		pkColumn, target, pkColumn)
	if err := conn.QueryRowCtx(ctx, &cnt, query); err != nil {
		return 0, fmt.Errorf("主键重复检查失败: %w", err)
	}
	return cnt, nil
}
