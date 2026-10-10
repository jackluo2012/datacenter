package logic

import (
	"context"
	"fmt"
	"strings"

	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetObjectTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetObjectTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetObjectTagsLogic {
	return &GetObjectTagsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查对象标签画像：按 object_pk 查打标值，join tag_def 出 tag_key/tag_name
func (l *GetObjectTagsLogic) GetObjectTags(in *tag.ObjectTagsReq) (*tag.ObjectTagsReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.ObjectId <= 0 {
		return nil, fmt.Errorf("%w: object_id 不能为空", ErrInvalidParams)
	}
	pk := strings.TrimSpace(in.ObjectPk)
	if pk == "" {
		return nil, fmt.Errorf("%w: object_pk 不能为空", ErrInvalidParams)
	}
	rows, err := l.svcCtx.TagValueModel.ListByObjectPk(l.ctx, in.ObjectId, in.Beid, pk)
	if err != nil {
		l.Errorf("查询对象标签画像失败, object_id: %d, object_pk: %s, err: %v", in.ObjectId, pk, err)
		return nil, err
	}
	items := make([]*tag.TagValueItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, &tag.TagValueItem{
			TagKey:   r.TagKey,
			TagName:  r.TagName,
			TagValue: r.TagValue,
			TaggedAt: r.TaggedAt.Unix(),
		})
	}
	return &tag.ObjectTagsReply{Tags: items}, nil
}
