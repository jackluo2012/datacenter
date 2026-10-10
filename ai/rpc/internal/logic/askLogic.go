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

// AI 问数：语义检索(元数据+指标) → 命中候选 → (可选 LLM) 生成 SQL 与回答，全程审计
func (l *AskLogic) Ask(in *ai.AskReq) (*ai.AskReply, error) {
	start := time.Now()

	//1. question 非空校验
	question := strings.TrimSpace(in.Question)
	if question == "" {
		return nil, fmt.Errorf("question 不能为空")
	}

	maxCand := int(in.MaxCandidates)
	if maxCand <= 0 {
		maxCand = defaultMaxCandidates
	}
	if maxCand > maxCandidatesLimit {
		maxCand = maxCandidatesLimit
	}

	//2. 关键词提取（MVP 分词降级法：去停用词 → 整句 + 分词）
	keywords := ExtractKeywords(question)

	//3. 语义检索 + 打分排序：元数据实体 + 指标合并候选，截取 top max_candidates
	candidates := RankCandidates(l.searchCandidates(l.ctx, in.Beid, keywords, maxCand), keywords, maxCand)

	reply := &ai.AskReply{
		Candidates: candidates,
	}
	if reply.Candidates == nil {
		reply.Candidates = []*ai.Candidate{}
	}

	//4. 生成回答：LLM 可选路径（Config.Llm.Enabled），失败降级为纯检索，不返回 error
	if l.svcCtx.Config.Llm.Enabled {
		// TODO(联调): 需配置真实 LLM ApiKey 验证生成质量
		client := llm.NewClient(l.svcCtx.Config.Llm.BaseUrl, l.svcCtx.Config.Llm.ApiKey,
			l.svcCtx.Config.Llm.Model, l.svcCtx.HttpClient)
		content, err := client.ChatCompletion(l.ctx, buildAskPrompt(question, candidates))
		if err != nil {
			//降级：answer 写明降级原因 + 错误摘要，llm_enabled=false
			l.Errorf("LLM 调用失败已降级为纯检索: %v", err)
			reply.Answer = fmt.Sprintf("%s\n\n[LLM 调用失败已降级为纯检索] 错误摘要: %v", buildSearchSummary(candidates), err)
		} else {
			reply.LlmEnabled = true
			reply.GeneratedSql = extractSQLBlock(content)
			reply.Answer = content
		}
	} else {
		reply.Answer = buildSearchSummary(candidates)
	}
	reply.CostMs = time.Since(start).Milliseconds()

	//5. 审计：无论成败写 dc_ask_log；审计写失败仅记日志，不阻断返回
	askLog := &model.AskLog{
		Question:       truncateRunes(question, maxQuestionRunes),
		CandidatesJson: truncateBytes(marshalCandidates(candidates), maxTextBytes),
		GeneratedSql:   truncateBytes(reply.GeneratedSql, maxTextBytes),
		Answer:         truncateBytes(reply.Answer, maxTextBytes),
		LlmEnabled:     boolToInt64(reply.LlmEnabled),
		CostMs:         reply.CostMs,
		Beid:           in.Beid,
	}
	if _, err := l.svcCtx.AskLogModel.Insert(l.ctx, askLog); err != nil {
		l.Errorf("写 dc_ask_log 审计失败(不阻断返回): %v", err)
	}

	return reply, nil
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
