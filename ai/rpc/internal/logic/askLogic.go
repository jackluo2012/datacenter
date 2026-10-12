package logic

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"datacenter/ai/model"
	"datacenter/ai/rpc/ai"
	"datacenter/ai/rpc/internal/llm"
	"datacenter/ai/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultMaxCandidates = 5     //候选数缺省（与 proto 注释一致）
	maxCandidatesLimit   = 20    //候选数上限（防御异常大值）
	maxQuestionRunes     = 1024  //dc_ask_log.question 为 varchar(1024)
	maxTextBytes         = 60000 //dc_ask_log text 列按字节截断兜底（上限 65535）
)

type AskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AskLogic {
	return &AskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// askCoreResult Ask 主流程（检索→LLM 生成）的中间产物，Ask 与 AskAndRun 共用：
// AskAndRun 在此基础上追加执行结果，并把最终 answer/耗时回填后再落审计
type askCoreResult struct {
	question     string          //清洗后的问句
	candidates   []*ai.Candidate //语义检索候选（永不为 nil，空结果为空数组）
	llmEnabled   bool            //LLM 是否生成成功
	generatedSql string          //LLM 生成的 SQL（未启用/失败/被安全校验拦截为空）
	sqlGuardErr  string          //生成 SQL 未通过发送侧安全校验的原因（通过校验/未生成时为空）
	answer       string          //回答（含降级说明）
	costMs       int64           //本阶段耗时
}

// AI 问数：语义检索(元数据+指标) → 命中候选 → (可选 LLM) 生成 SQL 与回答，全程审计
func (l *AskLogic) Ask(in *ai.AskReq) (*ai.AskReply, error) {
	//1. question 非空校验
	question := strings.TrimSpace(in.Question)
	if question == "" {
		return nil, fmt.Errorf("question 不能为空")
	}

	//2. 复用主流程：检索 → 可选 LLM 生成（context_json 可选携带上一轮 SQL，支持多轮追问）
	core := l.askCore(question, in.Beid, normalizeMaxCandidates(in.MaxCandidates), in.ContextJson)

	reply := &ai.AskReply{
		Candidates:   core.candidates,
		LlmEnabled:   core.llmEnabled,
		GeneratedSql: core.generatedSql,
		Answer:       core.answer,
		CostMs:       core.costMs,
	}
	if reply.Candidates == nil {
		reply.Candidates = []*ai.Candidate{}
	}

	//3. 审计：无论成败写 dc_ask_log
	l.writeAskAudit(core, in.Beid)
	return reply, nil
}

// askCore 问数主体（检索→候选打分→可选 LLM 生成 SQL），Ask 与 AskAndRun 共用；
// contextJson 为可选多轮上下文（AskReq.context_json 原文，空串=首轮）；
// 检索失败降级、LLM 失败降级均为常见业务态，不返回 error
func (l *AskLogic) askCore(question string, beid int64, maxCand int, contextJson string) *askCoreResult {
	start := time.Now()

	//1. 关键词提取（MVP 分词降级法：去停用词 → 整句 + 分词）
	keywords := ExtractKeywords(question)

	//2. 语义检索 + 打分排序：元数据实体 + 指标合并候选，截取 top max_candidates
	candidates := RankCandidates(l.searchCandidates(l.ctx, beid, keywords, maxCand), keywords, maxCand)

	core := &askCoreResult{
		question:   question,
		candidates: candidates,
	}
	if core.candidates == nil {
		core.candidates = []*ai.Candidate{}
	}

	//3. 生成回答：LLM 可选路径（Config.Llm.Enabled），失败降级为纯检索，不返回 error
	if l.svcCtx.Config.Llm.Enabled {
		client := llm.NewClient(l.svcCtx.Config.Llm.BaseUrl, l.svcCtx.Config.Llm.ApiKey,
			l.svcCtx.Config.Llm.Model, l.svcCtx.HttpClient)
		//system 放角色+严格规则+few-shot，user 放候选上下文+多轮上下文+本轮问题
		content, err := client.ChatCompletionWithSystem(l.ctx, buildAskSystemPrompt(),
			buildAskUserPrompt(question, core.candidates, contextJson))
		if err != nil {
			//降级：answer 写明降级原因 + 错误摘要，llm_enabled=false
			l.Errorf("LLM 调用失败已降级为纯检索: %v", err)
			core.answer = fmt.Sprintf("%s\n\n[LLM 调用失败已降级为纯检索] 错误摘要: %v", buildSearchSummary(candidates), err)
		} else {
			core.llmEnabled = true
			core.answer = content
			//发送侧安全校验（执行侧 ide 安全闸另有二次防护）：生成 SQL 必须是单条 SELECT，
			//不通过则清空 generatedSql 并在回答中说明，AskAndRun 不会拿被拦截的 SQL 去执行
			sql := extractSQLBlock(content)
			if sql == "" {
				//模型按规则合规拒答（未输出 sql 块）：正常业务态，AskAndRun 会提示无 SQL 可执行
			} else if err := ValidateGeneratedSql(sql); err != nil {
				//写语句/多语句片段：发送侧拦截，清空 generatedSql，AskAndRun 不再触达 ide
				core.sqlGuardErr = err.Error()
				l.Errorf("LLM 生成 SQL 未通过安全校验已拦截: question=%s err=%v", question, err)
				core.answer += fmt.Sprintf("\n\n[安全校验未通过] 生成内容不是单条 SELECT 查询(%v)，已拦截不执行；请调整问法或完善候选元数据。", err)
			} else {
				core.generatedSql = sql
			}
		}
	} else {
		core.answer = buildSearchSummary(candidates)
	}
	core.costMs = time.Since(start).Milliseconds()
	return core
}

// writeAskAudit 落 dc_ask_log 审计（Ask 与 AskAndRun 共用）：写失败仅记日志，不阻断返回；
// AskAndRun 在执行完成后回填 core.answer（尾部带执行摘要）/core.costMs 再调用，执行信息随 answer 落审计
func (l *AskLogic) writeAskAudit(r *askCoreResult, beid int64) {
	askLog := &model.AskLog{
		Question:       truncateRunes(r.question, maxQuestionRunes),
		CandidatesJson: truncateBytes(marshalCandidates(r.candidates), maxTextBytes),
		GeneratedSql:   truncateBytes(r.generatedSql, maxTextBytes),
		Answer:         truncateBytes(r.answer, maxTextBytes),
		LlmEnabled:     boolToInt64(r.llmEnabled),
		CostMs:         r.costMs,
		Beid:           beid,
	}
	if _, err := l.svcCtx.AskLogModel.Insert(l.ctx, askLog); err != nil {
		l.Errorf("写 dc_ask_log 审计失败(不阻断返回): %v", err)
	}
}

// normalizeMaxCandidates 候选数归一化（Ask 与 AskAndRun 共用，纯函数）：<=0 取默认，超上限截断
func normalizeMaxCandidates(n int32) int {
	if n <= 0 {
		return defaultMaxCandidates
	}
	if int(n) > maxCandidatesLimit {
		return maxCandidatesLimit
	}
	return int(n)
}

// buildSearchSummary 无 LLM（或降级）时的检索摘要回答：命中候选清单 + 引导提示
func buildSearchSummary(candidates []*ai.Candidate) string {
	if len(candidates) == 0 {
		return "语义检索未命中任何候选（元数据/指标无匹配），请更换关键词，或先在元数据/指标服务完善登记。"
	}
	parts := make([]string, 0, len(candidates))
	for _, c := range candidates {
		parts = append(parts, fmt.Sprintf("[%s] %s(%s)", c.Type, c.Name, c.Detail))
	}
	return fmt.Sprintf("语义检索命中 %d 个候选：%s。如需自动生成 SQL，请在配置中启用 LLM。",
		len(candidates), strings.Join(parts, "、"))
}

// truncateRunes 按字符数截断（question 为 varchar(1024) 字符）
func truncateRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

// truncateBytes 按字节数截断（text 列上限 65535 字节，按 rune 边界回退避免截半个字符）
func truncateBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}

// boolToInt64 bool 转 tinyint
func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
