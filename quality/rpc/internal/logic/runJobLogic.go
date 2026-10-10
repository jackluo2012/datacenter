package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"datacenter/datasource/rpc/datasource"
	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RunJobLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunJobLogic {
	return &RunJobLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行任务（同步跑完规则并落结果）：
// 查任务 → 解析 rule_ids → 按数据源服务拿 DSN 建目标库连接 →
// 逐规则生成检查 SQL 实测 → 结果批量落库 → 回填任务最近状态
func (l *RunJobLogic) RunJob(in *quality.IdReq) (*quality.RunJobReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, model.ErrJobIdRequired
	}
	job, err := l.svcCtx.JobModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	// 解析规则id（去重、保序）
	ruleIds, err := ParseRuleIds(job.RuleIds)
	if err != nil {
		return nil, err
	}
	// 加载规则，任一规则缺失（已删/跨租户）即中止
	rules, err := l.loadRulesInOrder(in.Beid, ruleIds)
	if err != nil {
		return nil, err
	}
	// 库名参与表名拼接，先过白名单
	if job.DbName != "" && !ValidIdentifier(job.DbName) {
		return nil, fmt.Errorf("%w: db_name %q", ErrInvalidIdentifier, job.DbName)
	}
	// 目标库连接：job.datasource_id → 数据源服务取 DSN
	conn, err := l.targetConn(job.DatasourceId, in.Beid)
	if err != nil {
		return nil, err
	}

	// 逐规则执行：单规则失败不阻断整体，错误落入结果详情（整体状态按通过数收敛）
	runAt := time.Now()
	results := make([]*model.Result, 0, len(rules))
	var passed int64
	for _, rule := range rules {
		result := l.execRule(conn, job.Id, job.DbName, rule, runAt)
		passed += result.PassFlag // PassFlag 为 1 时恰好累加 1
		results = append(results, result)
	}
	if err := l.svcCtx.ResultModel.BatchInsert(l.ctx, results); err != nil {
		return nil, err
	}

	// 全部通过 success、全部不通过 failed、部分通过 partial
	status := model.JobStatusPartial
	switch {
	case passed == int64(len(rules)):
		status = model.JobStatusSuccess
	case passed == 0:
		status = model.JobStatusFailed
	}
	if err := l.svcCtx.JobModel.UpdateRunStatus(l.ctx, job.Id, in.Beid, status, runAt); err != nil {
		return nil, err
	}
	return &quality.RunJobReply{
		JobId:      job.Id,
		TotalRules: int64(len(rules)),
		Passed:     passed,
		LastStatus: status,
	}, nil
}

// loadRulesInOrder 按任务配置的 rule_ids 顺序加载规则
func (l *RunJobLogic) loadRulesInOrder(beid int64, ruleIds []int64) ([]*model.Rule, error) {
	rules, err := l.svcCtx.RuleModel.FindListByIds(l.ctx, beid, ruleIds)
	if err != nil {
		return nil, err
	}
	byId := make(map[int64]*model.Rule, len(rules))
	for _, r := range rules {
		byId[r.Id] = r
	}
	ordered := make([]*model.Rule, 0, len(ruleIds))
	var missing []int64
	for _, id := range ruleIds {
		r, ok := byId[id]
		if !ok {
			missing = append(missing, id)
			continue
		}
		ordered = append(ordered, r)
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: 缺失id %v", ErrRuleNotFound, missing)
	}
	return ordered, nil
}

// targetConn 从数据源服务获取 DSN 并建立目标库连接（内部调用，with_secret=true 才返回连接串）
func (l *RunJobLogic) targetConn(datasourceId, beid int64) (sqlx.SqlConn, error) {
	reply, err := l.svcCtx.DataSourceCli.GetDataSource(l.ctx, &datasource.GetDataSourceReq{
		Id:         datasourceId,
		Beid:       beid,
		WithSecret: true,
	})
	if err != nil {
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}
	if reply == nil || reply.Datasource == nil || reply.Dsn == "" {
		return nil, ErrDataSourceDsn
	}
	if reply.Datasource.Type != "mysql" {
		return nil, fmt.Errorf("%w: 质量检查仅支持 mysql 数据源，当前 %q", ErrDataSourceDsn, reply.Datasource.Type)
	}
	// DSN 已含库名；job.db_name 非空时由 BuildCheckSQL 在表名前加 `<db>.` 前缀
	return sqlx.NewMysql(reply.Dsn), nil
}

// execRule 执行单条规则：生成检查 SQL → 目标库实测 → 阈值判定。
// 任一环节失败均记为不通过并把原因写入 detail_json，不阻断其余规则
func (l *RunJobLogic) execRule(conn sqlx.SqlConn, jobId int64, dbName string, rule *model.Rule, runAt time.Time) *model.Result {
	result := &model.Result{
		JobId:    jobId,
		RuleId:   rule.Id,
		RuleType: rule.RuleType,
		RunAt:    runAt.UnixMilli(),
	}
	// 生成检查 SQL（标识符白名单校验 + 数值参数占位符化）
	query, args, err := BuildCheckSQL(rule.RuleType, dbName, rule.TargetTable, rule.TargetColumn, rule.ParamsJson)
	if err != nil {
		result.DetailJson = marshalDetail(map[string]any{"error": err.Error()})
		return result
	}
	// 目标库实测
	var checkValue int64
	if err := conn.QueryRowCtx(l.ctx, &checkValue, query, args...); err != nil {
		result.DetailJson = marshalDetail(map[string]any{"error": err.Error(), "sql": query})
		return result
	}
	result.CheckValue = strconv.FormatInt(checkValue, 10)
	// 阈值判定
	pass, detail, err := JudgeCheck(rule.RuleType, checkValue, rule.ParamsJson, rule.ThresholdJson)
	if err != nil {
		result.DetailJson = marshalDetail(map[string]any{"error": err.Error()})
		return result
	}
	if pass {
		result.PassFlag = 1
	}
	result.DetailJson = detail
	return result
}
