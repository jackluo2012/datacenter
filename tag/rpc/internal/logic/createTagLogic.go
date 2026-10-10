package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTagLogic {
	return &CreateTagLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建标签定义：校验对象存在 → tag_type 白名单 → rule 类型强校验 rule_sql
// （必填、select 开头、含 as object_id/as tag_value 两列、禁分号防注入）→ 唯一查重后落库
func (l *CreateTagLogic) CreateTag(in *tag.CreateTagReq) (*tag.TagDefReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.ObjectId <= 0 {
		return nil, fmt.Errorf("%w: object_id 不能为空", ErrInvalidParams)
	}
	tagKey := strings.TrimSpace(in.TagKey)
	if tagKey == "" {
		return nil, fmt.Errorf("%w: tag_key 不能为空", ErrInvalidParams)
	}
	if strings.TrimSpace(in.TagName) == "" {
		return nil, fmt.Errorf("%w: tag_name 不能为空", ErrInvalidParams)
	}
	if !ValidTagType(in.TagType) {
		return nil, ErrInvalidTagType
	}
	// 所属对象必须存在（带租户隔离）
	if _, err := l.svcCtx.ObjectModel.FindOne(l.ctx, in.ObjectId, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	// rule 类型标签必须提供可通过校验的规则 SQL；其余类型清空存档
	var ruleSql string
	if in.TagType == TagTypeRule {
		if err := ValidateRuleSql(in.RuleSql); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidRuleSql, err)
		}
		ruleSql = strings.TrimSpace(in.RuleSql)
	}

	id, err := l.svcCtx.TagDefModel.Create(l.ctx, &model.TagDef{
		ObjectId:    in.ObjectId,
		TagKey:      tagKey,
		TagName:     strings.TrimSpace(in.TagName),
		TagType:     in.TagType,
		RuleSql:     ruleSql,
		RefreshCron: strings.TrimSpace(in.RefreshCron),
		Status:      1, //新建默认启用
		Beid:        in.Beid,
	})
	if err != nil {
		if errors.Is(err, model.ErrDupTagKey) {
			return nil, ErrDupTagKey
		}
		l.Errorf("创建标签定义失败, object_id: %d, tag_key: %s, err: %v", in.ObjectId, in.TagKey, err)
		return nil, err
	}
	def, err := l.svcCtx.TagDefModel.FindOne(l.ctx, id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &tag.TagDefReply{Def: toPbTagDef(def)}, nil
}
