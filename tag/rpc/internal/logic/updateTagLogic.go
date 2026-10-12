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

type UpdateTagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTagLogic {
	return &UpdateTagLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新标签定义（全量更新）：存在性校验 → 字段校验（key/name 必填、类型白名单、
// rule 类型强校验 rule_sql）→ 所属对象存在性 → 换 key/对象时唯一查重 → 落库
func (l *UpdateTagLogic) UpdateTag(in *tag.UpdateTagReq) (*tag.TagDefReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.Id <= 0 || in.ObjectId <= 0 {
		return nil, fmt.Errorf("%w: id/object_id 不能为空", ErrInvalidParams)
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
	// 状态仅允许 0.停用/1.启用，防脏状态落库
	if in.Status != 0 && in.Status != 1 {
		return nil, ErrInvalidTagStatus
	}
	// 存在性校验（带租户隔离）
	if _, err := l.svcCtx.TagDefModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	// 所属对象必须存在（带租户隔离）
	obj, err := l.svcCtx.ObjectModel.FindOne(l.ctx, in.ObjectId, in.Beid)
	if err != nil {
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
	// 唯一查重：同对象同 key 已被其他定义占用则拒绝（uk_obj_key）
	dup, err := l.svcCtx.TagDefModel.FindOneByKey(l.ctx, in.ObjectId, tagKey, in.Beid)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}
	if dup != nil && dup.Id != in.Id {
		return nil, ErrDupTagKey
	}
	if err := l.svcCtx.TagDefModel.Update(l.ctx, &model.TagDef{
		Id:          in.Id,
		ObjectId:    in.ObjectId,
		TagKey:      tagKey,
		TagName:     strings.TrimSpace(in.TagName),
		TagType:     in.TagType,
		RuleSql:     ruleSql,
		RefreshCron: strings.TrimSpace(in.RefreshCron),
		Status:      in.Status,
		Beid:        in.Beid,
	}); err != nil {
		// 并发窗口撞唯一键统一按重复处理
		if model.IsDuplicateEntry(err) {
			return nil, ErrDupTagKey
		}
		l.Errorf("更新标签定义失败, id: %d, object_id: %d, tag_key: %s, err: %v", in.Id, in.ObjectId, in.TagKey, err)
		return nil, err
	}
	def, err := l.svcCtx.TagDefModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &tag.TagDefReply{Def: withObjectName(toPbTagDef(def), obj)}, nil
}
