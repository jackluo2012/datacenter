package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMaskRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMaskRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMaskRuleLogic {
	return &DeleteMaskRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteMaskRuleLogic) DeleteMaskRule(in *security.IdReq) (*security.IdReq, error) {
	// todo: add your logic here and delete this line

	return &security.IdReq{}, nil
}
