package logic

import (
	"context"

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

func (l *ListMaskRulesLogic) ListMaskRules(in *security.MaskRuleListReq) (*security.MaskRuleListReply, error) {
	// todo: add your logic here and delete this line

	return &security.MaskRuleListReply{}, nil
}
