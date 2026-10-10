package logic

import (
	"context"

	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRuleLogic {
	return &DeleteRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除规则
func (l *DeleteRuleLogic) DeleteRule(in *quality.IdReq) (*quality.IdReq, error) {
	// todo: add your logic here and delete this line

	return &quality.IdReq{}, nil
}
