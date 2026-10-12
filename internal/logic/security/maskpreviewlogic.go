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

type MaskPreviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 脱敏预览
func NewMaskPreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MaskPreviewLogic {
	return &MaskPreviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 脱敏预览: 网关请求逐字段映射为 rpc 请求, rpc 侧按规则算法对样例值执行脱敏后返回结果。
// 契约差异: 网关 MaskPreviewResp 多带 rule_code, 取自 rpc MaskPreviewReply.RuleCode 便于前端核对命中规则。
func (l *MaskPreviewLogic) MaskPreview(req *types.MaskPreviewReq) (resp *types.MaskPreviewResp, err error) {
	reply, err := l.svcCtx.SecurityRpc.MaskPreview(l.ctx, &securityClient.MaskPreviewReq{
		RuleId: req.RuleId,
		Sample: req.Sample,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.MaskPreviewResp{
		Masked:   reply.GetMasked(),
		RuleCode: reply.GetRuleCode(),
	}, nil
}
