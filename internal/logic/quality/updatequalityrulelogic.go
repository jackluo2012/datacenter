// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package quality

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/quality/rpc/qualityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateQualityRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新质量规则
func NewUpdateQualityRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateQualityRuleLogic {
	return &UpdateQualityRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateQualityRule 转发 quality rpc 更新规则（全量更新，类型/参数/标识符校验在 rpc 侧），
// status 0.停用/1.启用 为必传契约，owner 未传时 rpc 侧保留原负责人
func (l *UpdateQualityRuleLogic) UpdateQualityRule(req *types.UpdateQualityRuleReq) (resp *types.QualityRule, err error) {
	reply, err := l.svcCtx.QualityRpc.UpdateRule(l.ctx, &qualityClient.UpdateRuleReq{
		Id:            req.Id,
		Name:          req.Name,
		RuleType:      req.RuleType,
		TargetTable:   req.TargetTable,
		TargetColumn:  req.TargetColumn,
		ParamsJson:    req.ParamsJson,
		ThresholdJson: req.ThresholdJson,
		Status:        req.Status,
		Owner:         req.Owner,
		Beid:          req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return qualityRuleToTypes(reply.GetRule()), nil
}
