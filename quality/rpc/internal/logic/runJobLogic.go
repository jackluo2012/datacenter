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
	// 失败规则及其结果（供错误样本抓取与告警明细两个钩子使用）
	failedRules := make([]*model.Rule, 0)
	failedResults := make([]*model.Result, 0)
	var passed int64
	for _, rule := range rules {
		result := l.execRule(conn, job.Id, job.DbName, rule, runAt)
		passed += result.PassFlag // PassFlag 为 1 时恰好累加 1
		results = append(results, result)
		if result.PassFlag == 0 {
			failedRules = append(failedRules, rule)
			failedResults = append(failedResults, result)
		}
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

	// 钩子一：not_null/unique/value_range 失败规则抓错误行样本落库（失败仅记日志，不阻断返回）
	l.captureErrorSamples(conn, job.Id, job.DbName, failedRules, runAt)
	// 钩子二：任务异常（failed/partial）按通道触发级别分发 webhook 告警（失败仅记日志，不阻断返回）
	l.dispatchAlerts(job, status, int64(len(rules)), passed, failedRules, failedResults)

	return &quality.RunJobReply{
		JobId:      job.Id,
		TotalRules: int64(len(rules)),
		Passed:     passed,
		LastStatus: status,
	}, nil
}

// captureErrorSamples 错误样本钩子：对 not_null/unique/value_range 失败规则抓取违规行样本
// （每条 limit 100）落 dc_quality_error_data，供 ListErrorData 下钻；
// 样本 SQL 生成失败/查询失败仅记日志，不影响任务结果返回
func (l *RunJobLogic) captureErrorSamples(conn sqlx.SqlConn, jobId int64, dbName string,
	failedRules []*model.Rule, runAt time.Time) {
	for _, rule := range failedRules {
		if err := captureOne(l.ctx, conn, l.svcCtx.ErrorDataModel, jobId, dbName, rule, runAt); err != nil {
			l.Errorf("错误样本抓取失败: job=%d rule=%d err=%v", jobId, rule.Id, err)
		}
	}
}

// dispatchAlerts 告警钩子：任务异常（failed/partial）时查询租户下启用的告警通道，
// 按 min_level 过滤后逐个推送 webhook（钉钉/飞书格式，钉钉 secret 加签可选），
// 发送结果逐条落 dc_quality_alert_log；通道查询/发送/落日志任一失败仅记日志，不阻断主流程
func (l *RunJobLogic) dispatchAlerts(job *model.Job, status string, total, passed int64,
	failedRules []*model.Rule, failedResults []*model.Result) {
	if status == model.JobStatusSuccess {
		return // 全部通过不告警
	}
	channels, err := l.svcCtx.AlertChannelModel.FindEnabled(l.ctx, job.Beid)
	if err != nil {
		l.Errorf("查询告警通道失败: job=%d beid=%d err=%v", job.Id, job.Beid, err)
		return
	}
	if len(channels) == 0 {
		return
	}
	// 失败规则摘要（告警内容：任务名/状态/通过率/失败规则明细）
	items := make([]alertFailedRule, 0, len(failedRules))
	for i := range failedRules {
		items = append(items, alertFailedRule{
			Name:     failedRules[i].Name,
			RuleType: failedRules[i].RuleType,
			Detail:   failedResults[i].DetailJson,
		})
	}
	content := BuildAlertContent(job.Name, job.Id, total, passed, status, items)
	for _, ch := range channels {
		// 触发级别过滤：min_level=failed 仅任务全部不通过才报；min_level=partial 部分失败即报
		if !LevelAllowed(ch.MinLevel, status) {
			continue
		}
		sendErr := sendAlertToChannel(ch, content)
		logRow := &model.AlertLog{
			ChannelId: ch.Id,
			JobId:     job.Id,
			Content:   content,
			Status:    model.AlertSendSuccess,
			SentAt:    time.Now().UnixMilli(),
		}
		if sendErr != nil {
			l.Errorf("告警发送失败: job=%d channel=%d(%s) err=%v", job.Id, ch.Id, ch.Name, sendErr)
			logRow.Status = model.AlertSendFailed
			logRow.Msg = sendErr.Error()
		}
		if err := l.svcCtx.AlertLogModel.Insert(l.ctx, logRow); err != nil {
			l.Errorf("告警日志落库失败: job=%d channel=%d err=%v", job.Id, ch.Id, err)
		}
	}
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
