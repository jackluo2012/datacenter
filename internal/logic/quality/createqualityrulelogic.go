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

type CreateQualityRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建质量规则
func NewCreateQualityRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateQualityRuleLogic {
	return &CreateQualityRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 创建质量规则: 网关请求逐字段映射为 rpc 请求, 转发后把规则转为网关类型返回
func (l *CreateQualityRuleLogic) CreateQualityRule(req *types.CreateQualityRuleReq) (resp *types.QualityRule, err error) {
	reply, err := l.svcCtx.QualityRpc.CreateRule(l.ctx, &qualityClient.CreateRuleReq{
		Name:          req.Name,
		RuleType:      req.RuleType,
		TargetTable:   req.TargetTable,
		TargetColumn:  req.TargetColumn,
		ParamsJson:    req.ParamsJson,
		ThresholdJson: req.ThresholdJson,
		Owner:         req.Owner,
		Beid:          req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return qualityRuleToTypes(reply.GetRule()), nil
}
