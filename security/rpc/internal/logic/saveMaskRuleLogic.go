package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveMaskRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveMaskRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveMaskRuleLogic {
	return &SaveMaskRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 脱敏规则
func (l *SaveMaskRuleLogic) SaveMaskRule(in *security.SaveMaskRuleReq) (*security.MaskRuleReply, error) {
	// todo: add your logic here and delete this line

	return &security.MaskRuleReply{}, nil
}
