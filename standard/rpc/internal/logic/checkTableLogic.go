package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"datacenter/datasource/rpc/datasource"
	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CheckTableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckTableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckTableLogic {
	return &CheckTableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// columnRow information_schema.COLUMNS 行扫描结构（与查询列 6/6 严格对齐，
// 规避 go-zero 少列查询扫宽 struct 报 not matching destination to scan）
type columnRow struct {
	ColumnName   string        `db:"column_name"`
	DataType     string        `db:"data_type"`
	ColumnType   string        `db:"column_type"`
	CharLen      sql.NullInt64 `db:"character_maximum_length"`
	NumPrecision sql.NullInt64 `db:"numeric_precision"`
	NumScale     sql.NullInt64 `db:"numeric_scale"`
}

// 标准核对：模型须已发布 → 数据源服务取 DSN 建目标库连接 →
// information_schema 读实际列 → 与标准模型 elements 逐一比对 → 差异落报告返回
func (l *CheckTableLogic) CheckTable(in *standard.CheckReq) (*standard.CheckReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.ModelId <= 0 {
		return nil, model.ErrModelIdRequired
	}
	if in.DatasourceId <= 0 {
		return nil, ErrDatasourceRequired
	}
	tableName := in.TableName
	if tableName == "" {
		return nil, ErrTableNameRequired
	}
	// 库名/表名参与 information_schema 查询条件，先过白名单
	if !ValidIdentifier(tableName) {
		return nil, fmt.Errorf("%w: table_name %q", ErrInvalidIdentifier, tableName)
	}
	if in.DbName != "" && !ValidIdentifier(in.DbName) {
		return nil, fmt.Errorf("%w: db_name %q", ErrInvalidIdentifier, in.DbName)
	}

	// 1.模型必须已发布（draft 定义随时在变，无核对意义）
	m, err := l.svcCtx.StdModelModel.FindOne(l.ctx, in.ModelId, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}
	if m.Status != model.ModelStatusPublished {
		return nil, ErrModelNotPublished
	}
	elements, err := ParseElementsJson(m.ElementsJson)
	if err != nil {
		return nil, err
	}
	if len(elements) == 0 {
		return nil, ErrElementsJsonInvalid
	}
	// 引用的数据元全部存在，并取定义用于类型比对
	byCode, err := resolveElements(l.ctx, l.svcCtx.ElementModel, in.Beid, elements)
	if err != nil {
		return nil, err
	}

	// 2.目标库连接与核对库名
	conn, dbName, err := l.targetConn(in.DatasourceId, in.Beid, in.DbName)
	if err != nil {
		return nil, err
	}

	// 3.读实际列：information_schema.COLUMNS 按库名+表名过滤（占位符传参，无拼接注入面）
	var rows []*columnRow
	query := "select column_name, data_type, column_type, character_maximum_length, " +
		"numeric_precision, numeric_scale from information_schema.COLUMNS " +
		"where table_schema = ? and table_name = ? order by ordinal_position"
	if err := conn.QueryRowsCtx(l.ctx, &rows, query, dbName, tableName); err != nil {
		return nil, fmt.Errorf("读取目标表结构失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: %s.%s", ErrTargetTableMissing, dbName, tableName)
	}
	actuals := make([]*ActualColumn, 0, len(rows))
	for _, r := range rows {
		actuals = append(actuals, &ActualColumn{
			Name:         r.ColumnName,
			DataType:     r.DataType,
			ColumnType:   r.ColumnType,
			CharLen:      r.CharLen.Int64,
			NumPrecision: r.NumPrecision.Int64,
			NumScale:     r.NumScale.Int64,
		})
	}

	// 4.逐一比对：每个 element_code 期望一列，类型按 typemap 规则映射匹配
	stdCols := make([]*StdColumn, 0, len(elements))
	for _, e := range elements {
		def := byCode[e.ElementCode]
		stdCols = append(stdCols, &StdColumn{
			Name:      def.ElementCode,
			DataType:  def.DataType,
			Length:    def.DataLength,
			Precision: def.DataPrecision,
		})
	}
	diff, pass := BuildStructDiff(stdCols, actuals)

	// 5.差异落 dc_std_check_report（pass_flag 1.符合 0.存在偏差）
	runAt := time.Now()
	passFlag := int64(0)
	if pass {
		passFlag = 1
	}
	reportId, err := l.svcCtx.CheckReportModel.Insert(l.ctx, &model.CheckReport{
		ModelId:      m.Id,
		DatasourceId: in.DatasourceId,
		DbName:       dbName,
		TableName:    tableName,
		DiffJson:     marshalDiff(diff),
		PassFlag:     passFlag,
		RunAt:        runAt.UnixMilli(),
		Beid:         in.Beid,
	})
	if err != nil {
		return nil, err
	}
	l.Infof("标准核对完成: model_id=%d table=%s.%s pass=%v diff=%d 项",
		m.Id, dbName, tableName, pass, len(diff))
	return &standard.CheckReply{Report: &standard.StdCheckReport{
		Id:        reportId,
		ModelId:   m.Id,
		TableName: tableName,
		DiffJson:  marshalDiff(diff),
		PassFlag:  pass,
		RunAt:     runAt.UnixMilli(),
	}}, nil
}

// targetConn 数据源服务取 DSN 并建立目标库连接，返回连接与核对用库名
// （CheckReq.db_name 优先；缺省时取 DSN 内默认库）
func (l *CheckTableLogic) targetConn(datasourceId, beid int64, dbName string) (sqlx.SqlConn, string, error) {
	reply, err := l.svcCtx.DataSourceRpc.GetDataSource(l.ctx, &datasource.GetDataSourceReq{
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
		return nil, "", fmt.Errorf("%w: 标准核对仅支持 mysql 数据源，当前 %q", ErrDataSourceDsn, reply.Datasource.Type)
	}
	// db_name 缺省时从 DSN 解析默认库
	if dbName == "" {
		cfg, perr := mysqldriver.ParseDSN(reply.Dsn)
		if perr != nil || cfg.DBName == "" {
			return nil, "", fmt.Errorf("%w: DSN 缺少默认库名且未指定 db_name", ErrDataSourceDsn)
		}
		dbName = cfg.DBName
	}
	conn, err := l.svcCtx.TargetConn(reply.Dsn)
	if err != nil {
		return nil, "", err
	}
	return conn, dbName, nil
}

// marshalDiff 差异项序列化为 diff_json（序列化失败降级为空串，不阻断报告落库）
func marshalDiff(items []*DiffItem) string {
	b, err := json.Marshal(items)
	if err != nil {
		return ""
	}
	return string(b)
}
