package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"datacenter/ai/rpc/ai"
	"datacenter/ai/rpc/internal/svc"
	"datacenter/ide/rpc/ideClient"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultRunRows  = 20        //执行返回行数缺省上限（与 proto 注释一致）
	maxRunRowsLimit = 100       //行数入参上限防御（ide 侧另有仅 SELECT 与保护性截断）
	runScriptPrefix = "ai-gen-" //AI 生成 SQL 落为临时脚本的名称前缀
)

type AskAndRunLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAskAndRunLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AskAndRunLogic {
	return &AskAndRunLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 问数并执行：Ask 全流程复用，auto_run=true 时把生成 SQL 先存为临时脚本，
// 再经 ide-rpc 安全闸（仅 SELECT、行数截断）真实执行并返回数据；
// 执行失败（SQL 不合法/数据源不可用等）是常见业务态：降级写入 run_error，不返回 error
func (l *AskAndRunLogic) AskAndRun(in *ai.AskAndRunReq) (*ai.AskAndRunReply, error) {
	start := time.Now()

	//1. 入参校验：question 非空；auto_run=true 时 datasource_id 必填（AI 无法确定生成 SQL 的目标库），
	//   beid 必填（ide 脚本存取与执行均按租户隔离）
	question := strings.TrimSpace(in.Question)
	if question == "" {
		return nil, fmt.Errorf("question 不能为空")
	}
	if in.AutoRun {
		if in.DatasourceId <= 0 {
			return nil, fmt.Errorf("auto_run=true 时 datasource_id 必填：AI 生成的 SQL 目标库未知，需调用方指定执行数据源")
		}
		if in.Beid <= 0 {
			return nil, fmt.Errorf("auto_run=true 时 beid 必填：脚本保存与执行均按租户隔离")
		}
	}

	//2. 复用 Ask 主流程（检索→候选打分→可选 LLM 生成 SQL；AskAndRun 暂不支持多轮，context_json 传空）
	ask := NewAskLogic(l.ctx, l.svcCtx)
	core := ask.askCore(question, in.Beid, normalizeMaxCandidates(in.MaxCandidates), "")

	reply := &ai.AskAndRunReply{
		Candidates:   core.candidates,
		LlmEnabled:   core.llmEnabled,
		GeneratedSql: core.generatedSql,
		Answer:       core.answer,
	}

	//3. 自动执行（auto_run=true 才走）：LLM 生成成功且有 SQL 才真实执行，否则说明原因写 run_error
	if in.AutoRun {
		switch {
		case !core.llmEnabled:
			reply.RunError = "未启用 LLM 或 LLM 调用失败已降级，无生成 SQL 可执行"
		case core.sqlGuardErr != "":
			//发送侧安全校验已拦截（仅允许单条 SELECT）：不送 ide 执行，原因透出
			reply.RunError = "生成 SQL 未通过安全校验: " + core.sqlGuardErr
		case strings.TrimSpace(core.generatedSql) == "":
			reply.RunError = "LLM 未生成 SQL（回复中未提取到 ```sql 代码块），无可执行内容"
		default:
			run, err := runGeneratedSqlViaIde(l.ctx, l.svcCtx.IdeRpc,
				core.generatedSql, in.Beid, in.DatasourceId, normalizeRunRows(in.MaxRows))
			if err != nil {
				//降级：run_error 带原因，不返回 error
				reply.RunError = err.Error()
				l.Errorf("AI 生成 SQL 自动执行失败已降级: question=%s err=%v", question, err)
			} else {
				reply.Executed = true
				reply.ColumnsJson = run.ColumnsJson
				reply.RowsJson = run.RowsJson
				reply.RowCount = run.RowCount
			}
		}
		//执行摘要追加到 answer 尾部：不改审计表，dc_ask_log.answer 随之记录 executed 与 row_count
		reply.Answer = appendRunSummary(reply.Answer, reply.Executed, reply.RunError, in.DatasourceId, reply.RowCount)
	}

	//4. 总耗时（含执行）+ 审计：回填最终 answer/耗时后复用 Ask 的审计落库
	reply.CostMs = time.Since(start).Milliseconds()
	core.answer = reply.Answer
	core.costMs = reply.CostMs
	ask.writeAskAudit(core, in.Beid)

	return reply, nil
}

// runGeneratedSqlViaIde AI 生成 SQL 的执行链路（独立函数便于单测替换 ideClient.Ide 桩）：
// ide 的 RunScript 只认已保存脚本 id，而 AI 生成的 SQL 是临时语句，故先 SaveScript 落为临时脚本
// （名称 ai-gen-<时间戳>，datasource_id 取调用方指定值），再经 ide 安全闸真实执行
func runGeneratedSqlViaIde(ctx context.Context, ideRpc ideClient.Ide, sql string,
	beid, datasourceId int64, maxRows int32) (*ideClient.RunScriptReply, error) {
	if ideRpc == nil {
		return nil, fmt.Errorf("ide rpc 客户端未装配")
	}
	//1. 存为临时脚本：SQL 原样保存，合法性由 ide 执行前的安全闸把关
	saveReply, err := ideRpc.SaveScript(ctx, &ideClient.SaveScriptReq{
		Name:         runScriptPrefix + time.Now().Format("20060102150405"),
		SqlText:      sql,
		DatasourceId: datasourceId,
		Beid:         beid,
	})
	if err != nil {
		return nil, fmt.Errorf("保存 AI 生成脚本失败: %w", err)
	}
	if saveReply.GetScript() == nil || saveReply.GetScript().GetId() <= 0 {
		return nil, fmt.Errorf("保存 AI 生成脚本失败: 未返回脚本 id")
	}
	//2. 经 ide 安全闸执行（仅放行单条 SELECT，行数保护截断）
	runReply, err := ideRpc.RunScript(ctx, &ideClient.RunScriptReq{
		Id:      saveReply.GetScript().GetId(),
		Beid:    beid,
		MaxRows: maxRows,
	})
	if err != nil {
		return nil, fmt.Errorf("ide 执行失败(安全闸仅放行单条 SELECT): %w", err)
	}
	return runReply, nil
}

// appendRunSummary 回答尾部附加执行摘要（纯函数便于单测）：
// 成功摘要含数据源与行数；未执行/失败摘要带 run_error 原因与排查提示
func appendRunSummary(answer string, executed bool, runError string, datasourceId, rowCount int64) string {
	var b strings.Builder
	b.WriteString(answer)
	if executed {
		fmt.Fprintf(&b, "\n\n[自动执行] 生成 SQL 已在数据源 %d 经安全闸执行，返回 %d 行（完整结果见 columns_json/rows_json）。",
			datasourceId, rowCount)
		return b.String()
	}
	if runError != "" {
		fmt.Fprintf(&b, "\n\n[自动执行未成功] 原因: %s。请检查生成 SQL 是否为单条 SELECT、目标数据源 %d 配置是否可用。",
			runError, datasourceId)
	}
	return b.String()
}

// normalizeRunRows 执行行数归一化（纯函数）：<=0 取默认 20，超上限截断（ide 侧另有保护性截断）
func normalizeRunRows(n int32) int32 {
	if n <= 0 {
		return defaultRunRows
	}
	if n > maxRunRowsLimit {
		return maxRunRowsLimit
	}
	return n
}
