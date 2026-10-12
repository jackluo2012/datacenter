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

type SecMaskRuleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 脱敏规则列表
func NewSecMaskRuleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecMaskRuleListLogic {
	return &SecMaskRuleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 脱敏规则列表: 页码默认值处理后转发, 业务错误直接透传。
// 空列表仍 make 出空切片, 保证 JSON 序列化为 [] 而非 null。
func (l *SecMaskRuleListLogic) SecMaskRuleList(req *types.MaskRuleListReq) (resp *types.MaskRuleListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.SecurityRpc.ListMaskRules(l.ctx, &securityClient.MaskRuleListReq{
		Beid: req.Beid,
		Page: page,
		Size: size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.SecMaskRule, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *maskRuleToTypes(item))
	}
	return &types.MaskRuleListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
