package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/mdm/model"
	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type DistributeRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDistributeRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DistributeRecordLogic {
	return &DistributeRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DistributeRecord 分发记录到目标数据源（真实写入目标表）：
//  1. 记录须 published（未发布拒绝）；
//  2. datasource rpc 换目标 DSN（tag 服务同款模式，with_secret=true）；
//  3. 目标表不存在则按模型 fields_json 自动建表（类型映射 string→varchar(255)/int→bigint/
//     decimal→decimal(18,4)/bool→tinyint(1)/date→date）；
//  4. replace into 幂等写入（重复分发整行替换）；
//  5. 成功/失败均落 dc_mdm_distribution——分发执行失败不返回 error，
//     返回带 failed 状态与失败原因 msg 的 DistributeReply
func (l *DistributeRecordLogic) DistributeRecord(in *mdm.DistributeReq) (*mdm.DistributeReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.RecordId <= 0 || in.TargetDsId <= 0 {
		return nil, ErrInvalidParams
	}

	//1. 记录须存在且已发布
	rec, err := l.svcCtx.RecordModel.FindOne(l.ctx, in.RecordId, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	if rec.Status != RecordStatusPublished {
		return nil, ErrRecordNotPublished
	}

	//2. 模型字段定义（自动建表列与值提取均以此为准）
	mdl, err := l.svcCtx.ModelModel.FindOne(l.ctx, rec.ModelId, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}
	fields, err := ParseFieldsJson(mdl.FieldsJson)
	if err != nil {
		return nil, err
	}

	//3. 目标库/表名规范化：表名为空取模型物理表名，标识符校验（防注入，库名表名会真实进 DDL）
	targetDb := trimSpace(in.TargetDb)
	targetTable := trimSpace(in.TargetTable)
	if targetTable == "" {
		targetTable = mdl.TableName
	}
	// target_db 可选（空=DSN 默认库），非空时才做标识符校验；target_table 必填
	if (targetDb != "" && (!ValidFieldName(targetDb) || len(targetDb) > targetIdentifierMaxLen)) ||
		!ValidFieldName(targetTable) || len(targetTable) > targetIdentifierMaxLen {
		return nil, ErrInvalidIdentifier
	}

	dist := &model.DistributionRow{
		RecordId:    rec.Id,
		TargetDsId:  in.TargetDsId,
		TargetDb:    targetDb,
		TargetTable: targetTable,
		Beid:        in.Beid,
	}

	//4. 执行分发（换 DSN → 自动建表 → replace into）；失败归入 failed 状态而非返回 error
	if err := l.doDistribute(fields, rec, in.TargetDsId, targetDb, targetTable); err != nil {
		dist.Status = model.DistStatusFailed
		dist.Msg = err.Error()
		l.Errorf("[DistributeRecord] 分发失败 record=%d -> ds=%d %s.%s: %v",
			rec.Id, in.TargetDsId, targetDb, targetTable, err)
	} else {
		dist.Status = model.DistStatusSuccess
		dist.Msg = "分发成功"
		l.Infof("[DistributeRecord] 分发成功 record=%d biz_code=%s -> ds=%d %s.%s beid=%d",
			rec.Id, rec.BizCode, in.TargetDsId, targetDb, targetTable, in.Beid)
	}

	//5. 落分发结果（成功/失败均留痕）
	id, err := l.svcCtx.DistributionModel.Insert(l.ctx, dist)
	if err != nil {
		return nil, err
	}
	dist.Id = id
	dist.DistributedAt = time.Now().UnixMilli()
	return &mdm.DistributeReply{Distribution: toPbDistribution(dist)}, nil
}

// doDistribute 执行分发写入：数据源换 DSN → 目标表不存在则按 fields_json 自动建表 → replace into 幂等写入
func (l *DistributeRecordLogic) doDistribute(fields []FieldDef, rec *model.RecordRow,
	targetDsId int64, targetDb, targetTable string) error {
	//1. datasource rpc 换目标 DSN（with_secret=true 返回拼好的连接串）
	dsReply, err := l.svcCtx.DataSourceRpc.GetDataSource(l.ctx, &dataSourceClient.GetDataSourceReq{
		Id:         targetDsId,
		Beid:       rec.Beid,
		WithSecret: true,
	})
	if err != nil {
		return fmt.Errorf("%w: 数据源 %d: %v", ErrDatasource, targetDsId, err)
	}
	dsn := strings.TrimSpace(dsReply.GetDsn())
	if dsn == "" {
		return fmt.Errorf("%w: 数据源 %d 未返回 DSN", ErrDatasource, targetDsId)
	}

	//2. 目标库连接（ServiceContext 按 DSN 缓存连接池）
	conn, err := l.svcCtx.TargetConn(dsn)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatasource, err)
	}

	qualified := buildTargetTable(targetDb, targetTable)

	//3. 目标表不存在则按模型字段定义自动建表（create table if not exists 幂等）
	if _, err := conn.ExecCtx(l.ctx, buildCreateTableDDL(qualified, fields)); err != nil {
		return fmt.Errorf("%w: 自动建表 %s: %v", ErrDistributeExec, qualified, err)
	}

	//4. 组装行值：biz_code + 按模型字段顺序取值（缺失的非必填字段补 NULL）+ created_at
	var data map[string]any
	if err := json.Unmarshal([]byte(rec.DataJson), &data); err != nil {
		return fmt.Errorf("%w: 记录 data_json 解析失败: %v", ErrDistributeExec, err)
	}
	args := make([]any, 0, len(fields)+2)
	args = append(args, rec.BizCode)
	for _, f := range fields {
		if v, ok := data[f.Name]; ok {
			args = append(args, v)
		} else {
			args = append(args, nil)
		}
	}
	args = append(args, time.Now())

	//5. replace into：主键/唯一键冲突整行替换，重复分发幂等
	insertSql, _ := buildReplaceSQL(qualified, fields)
	if _, err := conn.ExecCtx(l.ctx, insertSql, args...); err != nil {
		return fmt.Errorf("%w: 写入 %s: %v", ErrDistributeExec, qualified, err)
	}
	return nil
}
