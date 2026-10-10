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

type QualityRuleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 质量规则列表
func NewQualityRuleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QualityRuleListLogic {
	return &QualityRuleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 质量规则列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *QualityRuleListLogic) QualityRuleList(req *types.QualityRuleListReq) (resp *types.QualityRuleListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.QualityRpc.ListRules(l.ctx, &qualityClient.RuleListReq{
		Beid:     req.Beid,
		RuleType: req.RuleType,
		Page:     page,
		Size:     size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.QualityRule, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *qualityRuleToTypes(item))
	}
	return &types.QualityRuleListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
