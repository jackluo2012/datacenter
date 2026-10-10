package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveMaskRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveMaskRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveMaskRuleLogic {
	return &SaveMaskRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存脱敏规则（id>0 更新、id=0 新建；rule_code 同租户查重；
// algorithm 白名单 mask/hash/replace/null，params_json 按算法结构校验）
func (l *SaveMaskRuleLogic) SaveMaskRule(in *security.SaveMaskRuleReq) (*security.MaskRuleReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	ruleCode := strings.TrimSpace(in.RuleCode)
	if ruleCode == "" {
		return nil, ErrRuleCodeRequired
	}
	ruleName := strings.TrimSpace(in.RuleName)
	if ruleName == "" {
		return nil, ErrRuleNameRequired
	}
	algorithm := strings.TrimSpace(in.Algorithm)
	if !IsValidAlgorithm(algorithm) {
		return nil, ErrAlgorithmInvalid
	}
	// params_json 结构校验（纯函数）：mask 需非负 prefix/suffix，replace 需 replace 键
	if _, err := ParseMaskParams(algorithm, in.ParamsJson); err != nil {
		return nil, err
	}

	data := &model.MaskRule{
		RuleCode:   ruleCode,
		RuleName:   ruleName,
		Algorithm:  algorithm,
		ParamsJson: in.ParamsJson,
		Beid:       in.Beid,
	}

	if in.Id > 0 {
		// 更新：先确认存在，再对编码查重（排除自身）
		old, err := l.svcCtx.MaskRuleModel.FindOne(l.ctx, in.Id, in.Beid)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrRuleNotFound
			}
			return nil, err
		}
		if dup, derr := l.svcCtx.MaskRuleModel.FindByCode(l.ctx, in.Beid, ruleCode); derr == nil && dup.Id != old.Id {
			return nil, ErrRuleCodeDuplicate
		} else if derr != nil && !errors.Is(derr, model.ErrNotFound) {
			return nil, derr
		}
		data.Id = old.Id
		if err := l.svcCtx.MaskRuleModel.Update(l.ctx, data); err != nil {
			return nil, err
		}
	} else {
		// 新建：编码查重后插入
		if _, err := l.svcCtx.MaskRuleModel.FindByCode(l.ctx, in.Beid, ruleCode); err == nil {
			return nil, ErrRuleCodeDuplicate
		} else if !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
		id, err := l.svcCtx.MaskRuleModel.Insert(l.ctx, data)
		if err != nil {
			return nil, err
		}
		data.Id = id
	}

	// 回读保证返回与库内一致
	rule, err := l.svcCtx.MaskRuleModel.FindOne(l.ctx, data.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &security.MaskRuleReply{Rule: ruleToPb(rule)}, nil
}
