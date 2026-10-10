package logic

import (
	"context"
	"errors"

	"datacenter/security/model"
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

// 删除脱敏规则：被实体打标引用时拒绝删除
func (l *DeleteMaskRuleLogic) DeleteMaskRule(in *security.IdReq) (*security.IdReq, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	rule, err := l.svcCtx.MaskRuleModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrRuleNotFound
		}
		return nil, err
	}
	// 引用检查：任一实体打标绑定该规则即拒绝删除
	refs, err := l.svcCtx.EntityTagModel.CountByRuleCode(l.ctx, in.Beid, rule.RuleCode)
	if err != nil {
		return nil, err
	}
	if refs > 0 {
		return nil, ErrRuleInUse
	}
	if err := l.svcCtx.MaskRuleModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	return &security.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
