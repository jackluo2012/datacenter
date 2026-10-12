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

type AskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// AI 问数
func NewAskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AskLogic {
	return &AskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AI 问数: 候选上限默认值处理(maxCandidates<=0 取 5)后转发, 转发后把候选列表逐字段映射为网关类型返回
func (l *AskLogic) Ask(req *types.AskReq) (resp *types.AskResp, err error) {
	maxCandidates := req.MaxCandidates
	if maxCandidates <= 0 {
		maxCandidates = 5
	}
	reply, err := l.svcCtx.AiRpc.Ask(l.ctx, &aiClient.AskReq{
		Question:      req.Question,
		Beid:          req.Beid,
		MaxCandidates: maxCandidates,
		ContextJson:   req.ContextJson,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	candidates := make([]types.Candidate, 0, len(reply.GetCandidates()))
	for _, item := range reply.GetCandidates() {
		candidates = append(candidates, types.Candidate{
			Type:    item.GetType(),
			Id:      item.GetId(),
			Name:    item.GetName(),
			Detail:  item.GetDetail(),
			Comment: item.GetComment(),
		})
	}
	return &types.AskResp{
		Candidates:   candidates,
		LlmEnabled:   reply.GetLlmEnabled(),
		GeneratedSql: reply.GetGeneratedSql(),
		Answer:       reply.GetAnswer(),
		CostMs:       reply.GetCostMs(),
	}, nil
}
