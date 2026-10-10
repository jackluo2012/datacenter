package logic

import (
	"context"

	"datacenter/quality/model"
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

// 删除规则（幂等：规则不存在时不报错）
func (l *DeleteRuleLogic) DeleteRule(in *quality.IdReq) (*quality.IdReq, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrInvalidRuleParams
	}
	if err := l.svcCtx.RuleModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	return &quality.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
