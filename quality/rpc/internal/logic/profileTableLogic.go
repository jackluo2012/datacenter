package logic

import (
	"context"
	"time"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileTableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProfileTableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileTableLogic {
	return &ProfileTableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 表剖析与健康分
// 流程：数据源服务取 DSN 建目标库连接 → 总行数 → 逐列剖析（空值数/去重数/最小最大样本）→
// 主键重复组数（可选）→ 健康分（纯函数）→ profile_json 逐列明细 → 落 dc_quality_profile → 返回
func (l *ProfileTableLogic) ProfileTable(in *quality.ProfileTableReq) (*quality.ProfileReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.DatasourceId <= 0 {
		return nil, ErrDatasourceRequired
	}
	// 表名必填，库名/主键列选填但均须过标识符白名单
	if in.TableName == "" {
		return nil, ErrTargetTableRequired
	}
	if !ValidIdentifier(in.TableName) {
		return nil, errInvalidIdentifier("table_name", in.TableName)
	}
	if in.DbName != "" && !ValidIdentifier(in.DbName) {
		return nil, errInvalidIdentifier("db_name", in.DbName)
	}
	if in.PkColumn != "" && !ValidIdentifier(in.PkColumn) {
		return nil, errInvalidIdentifier("pk_column", in.PkColumn)
	}

	// 目标库连接：datasource_id → 数据源服务取 DSN（与 RunJob 同装配模式）
	conn, dsn, err := profileTargetConn(l.ctx, l.svcCtx.DataSourceCli, in.DatasourceId, in.Beid)
	if err != nil {
		return nil, err
	}
	// 目标库名：请求显式指定优先，否则从 DSN 默认库解析
	dbName, err := ProfileDbName(in.DbName, dsn)
	if err != nil {
		return nil, err
	}
	if !ValidIdentifier(dbName) {
		return nil, errInvalidIdentifier("db_name", dbName)
	}
	target, err := QualifyTable(dbName, in.TableName)
	if err != nil {
		return nil, err
	}

	// 目标表存在性前置校验（information_schema 计数，给出可读的 NotFound）
	if ok, err := tableExists(l.ctx, conn, dbName, in.TableName); err != nil {
		return nil, err
	} else if !ok {
		return nil, model.ErrNotFound
	}

	// 总行数
	var rowsCnt int64
	if err := conn.QueryRowCtx(l.ctx, &rowsCnt, "select count(*) from "+target); err != nil {
		return nil, err
	}

	// 逐列剖析（单列一条聚合 SQL：空值数/去重数/最小最大样本，列名过白名单后拼入）
	cols, err := listTableColumns(l.ctx, conn, dbName, in.TableName)
	if err != nil {
		return nil, err
	}
	profiles := make([]*columnProfile, 0, len(cols))
	var nullCells int64
	for _, c := range cols {
		agg, err := profileColumnAgg(l.ctx, conn, target, c.Column)
		if err != nil {
			return nil, err
		}
		nullCells += agg.NullCnt
		profiles = append(profiles, &columnProfile{
			Column:      c.Column,
			NullCnt:     agg.NullCnt,
			NullRate:    ratio(agg.NullCnt, rowsCnt),
			DistinctCnt: agg.DistinctCnt,
			SampleMin:   agg.SampleMin,
			SampleMax:   agg.SampleMax,
		})
	}

	// 主键重复组数（传入 pk_column 才检查）
	var dupPkCnt int64
	if in.PkColumn != "" {
		if dupPkCnt, err = countDuplicateGroups(l.ctx, conn, target, in.PkColumn); err != nil {
			return nil, err
		}
	}

	// 健康分：100 - 空值率×40 - 主键重复占比×30 - 自定义扣分项（空表扣 10），夹取 0-100
	totalCells := rowsCnt * int64(len(cols))
	nullRate := ratio(nullCells, totalCells)
	extraDeduct := 0.0
	if rowsCnt == 0 {
		extraDeduct = emptyTableDeduct
	}
	healthScore := HealthScore(nullRate, ratio(dupPkCnt, rowsCnt), extraDeduct)

	// 落库并返回
	now := time.Now()
	ranAt := now.UnixMilli()
	profile := &model.Profile{
		DatasourceId: in.DatasourceId,
		DbName:       dbName,
		TableName:    in.TableName,
		RowsCnt:      rowsCnt,
		ColsCnt:      int64(len(cols)),
		NullCells:    nullCells,
		TotalCells:   totalCells,
		NullRate:     nullRate,
		DupPkCnt:     dupPkCnt,
		HealthScore:  int64(healthScore),
		ProfileJson:  BuildColumnProfileJSON(profiles),
		RanAt:        ranAt,
		Beid:         in.Beid,
	}
	id, err := l.svcCtx.ProfileModel.Insert(l.ctx, profile)
	if err != nil {
		return nil, err
	}
	profile.Id = id
	return &quality.ProfileReply{Profile: profileToPb(profile)}, nil
}
