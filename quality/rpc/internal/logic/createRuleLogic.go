package logic

import (
	"context"

	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRuleLogic {
	return &CreateRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建规则
func (l *CreateRuleLogic) CreateRule(in *quality.CreateRuleReq) (*quality.RuleReply, error) {
	// todo: add your logic here and delete this line

	return &quality.RuleReply{}, nil
}
