// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSecMaskRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteSecMaskRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSecMaskRuleLogic {
	return &DeleteSecMaskRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSecMaskRuleLogic) DeleteSecMaskRule(req *types.SecIdReq) error {
	// todo: add your logic here and delete this line

	return nil
}
