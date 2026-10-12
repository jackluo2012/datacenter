package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRuleLogic {
	return &UpdateRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新规则（全量更新）：存在性校验 → 名称/类型白名单/标识符白名单/按类型参数校验 →
// 状态仅允许 0/1 → owner 空则保留原值 → 落库后回查返回
func (l *UpdateRuleLogic) UpdateRule(in *quality.UpdateRuleReq) (*quality.RuleReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrInvalidRuleParams
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, ErrRuleNameRequired
	}
	// 状态白名单：仅 0.停用/1.启用，防脏状态落库
	if in.Status != model.RuleStatusDisabled && in.Status != model.RuleStatusEnabled {
		return nil, ErrInvalidRuleParams
	}
	// 类型白名单 + 标识符白名单 + 按类型的 target_column/params_json 必填校验（与创建同口径）
	if err := ValidateRule(in.RuleType, in.TargetTable, in.TargetColumn, in.ParamsJson); err != nil {
		return nil, err
	}
	// 存在性校验（带租户隔离），不存在/跨租户不可见直接拒绝
	old, err := l.svcCtx.RuleModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrRuleNotFound
		}
		return nil, err
	}
	// owner 为空保留原负责人，避免表单未带负责人时误清空
	owner := mergeOwner(in.Owner, old.Owner)
	if err := l.svcCtx.RuleModel.Update(l.ctx, &model.Rule{
		Id:            in.Id,
		Name:          strings.TrimSpace(in.Name),
		RuleType:      in.RuleType,
		TargetTable:   in.TargetTable,
		TargetColumn:  in.TargetColumn,
		ParamsJson:    in.ParamsJson,
		ThresholdJson: in.ThresholdJson,
		Status:        int64(in.Status),
		Owner:         owner,
		Beid:          in.Beid,
	}); err != nil {
		l.Errorf("更新规则失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, err
	}
	rule, err := l.svcCtx.RuleModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &quality.RuleReply{Rule: ruleToPb(rule)}, nil
}
