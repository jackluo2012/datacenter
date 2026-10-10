package logic

import (
	"context"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRulesLogic {
	return &ListRulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 规则列表（分页 + rule_type 可选过滤）
func (l *ListRulesLogic) ListRules(in *quality.RuleListReq) (*quality.RuleListReply, error) {
	// 多租户条件必填
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.RuleModel.FindList(l.ctx, in.Beid, in.RuleType, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	rules := make([]*quality.QualityRule, 0, len(list))
	for _, r := range list {
		rules = append(rules, ruleToPb(r))
	}
	return &quality.RuleListReply{List: rules, Total: total}, nil
}
