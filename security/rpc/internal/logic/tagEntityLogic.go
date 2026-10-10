package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type TagEntityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTagEntityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagEntityLogic {
	return &TagEntityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 实体分级打标（uk_beid_entity upsert，同实体重复打标等价覆盖）：
// level_code 必须在分级目录中；mask_rule_code 非空时规则必须存在
func (l *TagEntityLogic) TagEntity(in *security.TagEntityReq) (*security.EntityTagReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.EntityId <= 0 {
		return nil, ErrEntityIdRequired
	}
	levelCode := strings.TrimSpace(in.LevelCode)
	if !ValidLevelCode(levelCode) {
		return nil, ErrLevelCodeInvalid
	}
	// 密级必须在分级目录（含租户自定义级别，跨租户目录不可见）
	if _, err := l.svcCtx.LevelModel.FindByCode(l.ctx, in.Beid, levelCode); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrLevelRefNotFound
		}
		return nil, err
	}
	// 绑定的脱敏规则非空时必须存在
	maskRuleCode := strings.TrimSpace(in.MaskRuleCode)
	if maskRuleCode != "" {
		if _, err := l.svcCtx.MaskRuleModel.FindByCode(l.ctx, in.Beid, maskRuleCode); err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrMaskRuleRefNotFound
			}
			return nil, err
		}
	}

	id, err := l.svcCtx.EntityTagModel.TagEntity(l.ctx, &model.EntityTag{
		EntityId:     in.EntityId,
		EntityName:   in.EntityName,
		LevelCode:    levelCode,
		MaskRuleCode: maskRuleCode,
		Beid:         in.Beid,
	})
	if err != nil {
		return nil, err
	}

	// 回读保证返回与库内一致
	tag, err := l.svcCtx.EntityTagModel.FindOne(l.ctx, id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &security.EntityTagReply{Tag: tagToPb(tag)}, nil
}
