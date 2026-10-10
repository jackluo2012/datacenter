package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/theme/model"
	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

// ddlSummaryMaxLen 部署结果 msg 携带的生成 SQL 摘要长度上限
const ddlSummaryMaxLen = 300

type DeployThemeModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeployThemeModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeployThemeModelLogic {
	return &DeployThemeModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeployThemeModel 按模型在目标库真实建表：
//  1. 模型须存在，存量 fields_json/layer/表名/库名防御性复检（白名单+pk 恰有一个）；
//  2. 由 fields_json 生成幂等 DDL：create table if not exists <layer>_<table>
//     (<field> <sqltype>, ..., primary key(<pk>))，类型映射 string→varchar(255)/
//     int→int/bigint→bigint/decimal→decimal(18,4)/bool→tinyint(1)/date→date/
//     datetime→datetime(3)；
//  3. datasource rpc 换目标 DSN（mdm 服务同款模式，with_secret=true）；
//  4. 真实执行 DDL；
//  5. 返回 ok+msg（msg 含生成 SQL 摘要）；数据源/执行失败不返回 error，
//     以 ok=false + 失败原因 msg 回告（便于前端直接展示）。
func (l *DeployThemeModelLogic) DeployThemeModel(in *theme.IdReq) (*theme.RefreshReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}

	//1. 模型须存在
	m, err := l.svcCtx.ModelModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}
	//2. 存量定义防御性复检（保存时已把过关，此处防绕过直写库的脏数据）
	if !ValidLayer(m.Layer) {
		return nil, ErrInvalidLayer
	}
	if !ValidIdentifier(m.TableName) || len(m.TableName) > identifierMaxLen ||
		len(buildPhysicalTable(m.Layer, m.TableName)) > identifierMaxLen {
		return nil, ErrInvalidTableName
	}
	dbName := trimSpace(m.DbName)
	if dbName != "" && (!ValidIdentifier(dbName) || len(dbName) > identifierMaxLen) {
		return nil, ErrInvalidDbName
	}
	if m.DatasourceId <= 0 {
		return nil, fmt.Errorf("%w: 模型未绑定目标数据源", ErrInvalidParams)
	}
	fields, err := ValidateFieldsJson(m.FieldsJson)
	if err != nil {
		return nil, err
	}

	//3. 生成幂等建表 DDL（物理表名 <layer>_<table>，db_name 非空时限定库名）
	qualified := buildQualifiedTable(dbName, buildPhysicalTable(m.Layer, m.TableName))
	ddl, err := buildCreateTableDDL(qualified, fields)
	if err != nil {
		return nil, err
	}

	//4. 数据源换目标 DSN → 真实执行（失败落 ok=false + 原因 msg，不返回 error）
	start := time.Now()
	err = l.deploy(ddl, m.DatasourceId, in.Beid)
	costMs := time.Since(start).Milliseconds()
	if err != nil {
		l.Errorf("[DeployThemeModel] 部署失败 model=%d -> ds=%d %s: %v",
			m.Id, m.DatasourceId, qualified, err)
		return &theme.RefreshReply{Ok: false, Msg: err.Error(), CostMs: costMs}, nil
	}
	l.Infof("[DeployThemeModel] 部署成功 model=%d table=%s beid=%d cost=%dms ddl=%s",
		m.Id, qualified, in.Beid, costMs, summarizeSql(ddl, ddlSummaryMaxLen))
	return &theme.RefreshReply{
		Ok:     true,
		Msg:    fmt.Sprintf("模型 %s 部署成功，DDL: %s", qualified, summarizeSql(ddl, ddlSummaryMaxLen)),
		CostMs: costMs,
	}, nil
}

// deploy 执行建表：数据源换 DSN（with_secret=true 返回拼好的连接串）→ 目标库连接 → 执行 DDL
func (l *DeployThemeModelLogic) deploy(ddl string, datasourceId, beid int64) error {
	dsReply, err := l.svcCtx.DataSourceRpc.GetDataSource(l.ctx, &dataSourceClient.GetDataSourceReq{
		Id:         datasourceId,
		Beid:       beid,
		WithSecret: true,
	})
	if err != nil {
		return fmt.Errorf("%w: 数据源 %d: %v", ErrDatasource, datasourceId, err)
	}
	dsn := strings.TrimSpace(dsReply.GetDsn())
	if dsn == "" {
		return fmt.Errorf("%w: 数据源 %d 未返回 DSN", ErrDatasource, datasourceId)
	}
	//目标库连接（ServiceContext 按 DSN 缓存连接池）
	conn, err := l.svcCtx.TargetConn(dsn)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatasource, err)
	}
	if _, err := conn.ExecCtx(l.ctx, ddl); err != nil {
		return fmt.Errorf("%w: %v", ErrDeployExec, err)
	}
	return nil
}
