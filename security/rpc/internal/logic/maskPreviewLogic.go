package logic

import (
	"context"
	"errors"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type MaskPreviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMaskPreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MaskPreviewLogic {
	return &MaskPreviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 脱敏预览：按规则对样例值执行脱敏算法（纯逻辑不落库）。
// MaskPreviewReq 契约不含 beid，按主键查询不做租户过滤
func (l *MaskPreviewLogic) MaskPreview(in *security.MaskPreviewReq) (*security.MaskPreviewReply, error) {
	if in.RuleId <= 0 {
		return nil, ErrIdRequired
	}
	rule, err := l.svcCtx.MaskRuleModel.FindOne(l.ctx, in.RuleId, 0)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrRuleNotFound
		}
		return nil, err
	}
	// 规则落库时已校验过 params_json，此处再解析一次防脏数据
	params, err := ParseMaskParams(rule.Algorithm, rule.ParamsJson)
	if err != nil {
		return nil, err
	}
	return &security.MaskPreviewReply{
		Masked:   MaskValue(rule.Algorithm, in.Sample, params),
		RuleCode: rule.RuleCode,
	}, nil
}
