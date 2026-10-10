// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SecMaskRuleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSecMaskRuleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecMaskRuleListLogic {
	return &SecMaskRuleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SecMaskRuleListLogic) SecMaskRuleList(req *types.MaskRuleListReq) (resp *types.MaskRuleListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
