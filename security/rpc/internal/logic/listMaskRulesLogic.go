package logic

import (
	"context"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMaskRulesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMaskRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMaskRulesLogic {
	return &ListMaskRulesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 脱敏规则列表（分页）
func (l *ListMaskRulesLogic) ListMaskRules(in *security.MaskRuleListReq) (*security.MaskRuleListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.MaskRuleModel.List(l.ctx, in.Beid, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	pbList := make([]*security.MaskRule, 0, len(list))
	for _, rule := range list {
		pbList = append(pbList, ruleToPb(rule))
	}
	return &security.MaskRuleListReply{List: pbList, Total: total}, nil
}
