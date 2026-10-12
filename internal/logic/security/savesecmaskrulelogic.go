// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/security/rpc/securityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveSecMaskRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存脱敏规则
func NewSaveSecMaskRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSecMaskRuleLogic {
	return &SaveSecMaskRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存脱敏规则: 网关请求逐字段映射为 rpc 请求, 转发后把规则转为网关类型返回。
// 契约差异: proto SaveMaskRuleReq 定义了 Id(id>0 走更新语义), 但网关 SaveSecMaskRuleReq 无 Id 字段
// 无法透传, rpc 侧 Id 恒为 0, 即始终按新建处理。
func (l *SaveSecMaskRuleLogic) SaveSecMaskRule(req *types.SaveSecMaskRuleReq) (resp *types.SecMaskRule, err error) {
	reply, err := l.svcCtx.SecurityRpc.SaveMaskRule(l.ctx, &securityClient.SaveMaskRuleReq{
		RuleCode:   req.RuleCode,
		RuleName:   req.RuleName,
		Algorithm:  req.Algorithm,
		ParamsJson: req.ParamsJson,
		Beid:       req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return maskRuleToTypes(reply.GetRule()), nil
}
