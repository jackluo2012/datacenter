package logic

import (
	"context"
	"strings"

	"datacenter/quality/model"
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
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, ErrRuleNameRequired
	}
	// 类型白名单 + 标识符白名单 + 按类型的 target_column/params_json 必填校验
	if err := ValidateRule(in.RuleType, in.TargetTable, in.TargetColumn, in.ParamsJson); err != nil {
		return nil, err
	}
	id, err := l.svcCtx.RuleModel.Insert(l.ctx, &model.Rule{
		Name:          in.Name,
		RuleType:      in.RuleType,
		TargetTable:   in.TargetTable,
		TargetColumn:  in.TargetColumn,
		ParamsJson:    in.ParamsJson,
		ThresholdJson: in.ThresholdJson,
		Status:        model.RuleStatusEnabled,
		Owner:         in.Owner,
		Beid:          in.Beid,
	})
	if err != nil {
		return nil, err
	}
	rule, err := l.svcCtx.RuleModel.FindOne(l.ctx, id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &quality.RuleReply{Rule: ruleToPb(rule)}, nil
}
