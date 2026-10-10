// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package ai

import (
	"context"

	"datacenter/ai/rpc/aiClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AskAndRunLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// AI 问数并执行（LLM 生成 SQL 后经 ide 安全闸自动执行）
func NewAskAndRunLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AskAndRunLogic {
	return &AskAndRunLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AskAndRunLogic) AskAndRun(req *types.AskAndRunReq) (resp *types.AskAndRunResp, err error) {
	// max_rows 默认 20（rpc 侧上限 100）
	maxRows := req.MaxRows
	if maxRows <= 0 {
		maxRows = 20
	}
	// max_candidates 默认 5（与 Ask 口径一致）
	maxCandidates := req.MaxCandidates
	if maxCandidates <= 0 {
		maxCandidates = 5
	}

	reply, err := l.svcCtx.AiRpc.AskAndRun(l.ctx, &aiClient.AskAndRunReq{
		Question:      req.Question,
		Beid:          req.Beid,
		MaxCandidates: int32(maxCandidates),
		AutoRun:       req.AutoRun,
		DatasourceId:  req.DatasourceId,
		MaxRows:       int32(maxRows),
	})
	if err != nil {
		return nil, err //业务错误直接透传 rpc error
	}

	candidates := make([]types.Candidate, 0, len(reply.Candidates))
	for _, c := range reply.Candidates {
		candidates = append(candidates, types.Candidate{
			Type:    c.GetType(),
			Id:      c.GetId(),
			Name:    c.GetName(),
			Detail:  c.GetDetail(),
			Comment: c.GetComment(),
		})
	}

	return &types.AskAndRunResp{
		Candidates:   candidates,
		LlmEnabled:   reply.GetLlmEnabled(),
		GeneratedSql: reply.GetGeneratedSql(),
		Answer:       reply.GetAnswer(),
		CostMs:       reply.GetCostMs(),
		Executed:     reply.GetExecuted(),
		ColumnsJson:  reply.GetColumnsJson(),
		RowsJson:     reply.GetRowsJson(),
		RowCount:     reply.GetRowCount(),
		RunError:     reply.GetRunError(),
	}, nil
}
