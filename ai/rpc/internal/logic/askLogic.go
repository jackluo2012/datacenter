package logic

import (
	"context"

	"datacenter/ai/rpc/ai"
	"datacenter/ai/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
	// todo: add your logic here and delete this line

	return &ai.AskReply{}, nil
}
