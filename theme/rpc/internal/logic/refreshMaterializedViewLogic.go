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

type RefreshMaterializedViewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshMaterializedViewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshMaterializedViewLogic {
	return &RefreshMaterializedViewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RefreshMaterializedView 真实刷新物化视图：
//  1. 定义须存在，view_name/库名白名单复检 + select_sql 复检（SELECT 开头、无分号）；
//  2. datasource rpc 换目标 DSN（mdm 服务同款模式，with_secret=true）；
//  3. 真实执行 create or replace view <view> as <select_sql>（MySQL 普通视图即可
//     满足"物化"语义，create or replace 原子替换重建，等效 drop+create）；
//  4. 成功 last_status=success / 失败 last_status=failed，均回写 last_refresh_at；
//  5. 成功/失败均以 ok+msg 回告（不返回 error，便于前端直接展示与留痕）。
func (l *RefreshMaterializedViewLogic) RefreshMaterializedView(in *theme.IdReq) (*theme.RefreshReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}

	//1. 定义须存在
	mv, err := l.svcCtx.ThemeMvModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrMvNotFound
		}
		return nil, err
	}
	//2. 存量定义防御性复检（建档时已把过关，此处防绕过直写库的脏数据）
	fail := func(status string, cause error) (*theme.RefreshReply, error) {
		//失败亦留痕：last_status=failed + last_refresh_at
		if uerr := l.svcCtx.ThemeMvModel.UpdateRunStatus(l.ctx, mv.Id, mv.Beid, status); uerr != nil {
			return nil, uerr
		}
		l.Errorf("[RefreshMaterializedView] 刷新失败 mv=%d view=%s beid=%d: %v",
			mv.Id, mv.ViewName, mv.Beid, cause)
		return &theme.RefreshReply{Ok: false, Msg: cause.Error()}, nil
	}

	if !ValidIdentifier(mv.ViewName) || len(mv.ViewName) > identifierMaxLen {
		return fail(MvStatusFailed, ErrInvalidViewName)
	}
	dbName := trimSpace(mv.DbName)
	if dbName != "" && (!ValidIdentifier(dbName) || len(dbName) > identifierMaxLen) {
		return fail(MvStatusFailed, ErrInvalidDbName)
	}
	if mv.DatasourceId <= 0 {
		return fail(MvStatusFailed, fmt.Errorf("%w: 未绑定数据源", ErrInvalidParams))
	}
	if err := ValidateSelectSql(mv.SelectSql); err != nil {
		//select_sql 非法（须 SELECT 开头且无分号）：报错入 last_status=failed
		return fail(MvStatusFailed, err)
	}

	//3. 生成建视图 DDL 并换目标库执行
	qualified := buildQualifiedTable(dbName, mv.ViewName)
	ddl := buildCreateViewDDL(qualified, mv.SelectSql)
	start := time.Now()
	if err := l.exec(ddl, mv.DatasourceId, mv.Beid); err != nil {
		return fail(MvStatusFailed, err)
	}
	costMs := time.Since(start).Milliseconds()

	//4. 成功留痕：last_status=success + last_refresh_at
	if err := l.svcCtx.ThemeMvModel.UpdateRunStatus(l.ctx, mv.Id, mv.Beid, MvStatusSuccess); err != nil {
		return nil, err
	}
	l.Infof("[RefreshMaterializedView] 刷新成功 mv=%d view=%s beid=%d cost=%dms sql=%s",
		mv.Id, qualified, mv.Beid, costMs, summarizeSql(ddl, ddlSummaryMaxLen))
	return &theme.RefreshReply{
		Ok:     true,
		Msg:    fmt.Sprintf("物化视图 %s 刷新成功，SQL: %s", qualified, summarizeSql(ddl, ddlSummaryMaxLen)),
		CostMs: costMs,
	}, nil
}

// exec 执行建视图：数据源换 DSN（with_secret=true 返回拼好的连接串）→ 目标库连接 → 执行 DDL
func (l *RefreshMaterializedViewLogic) exec(ddl string, datasourceId, beid int64) error {
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
