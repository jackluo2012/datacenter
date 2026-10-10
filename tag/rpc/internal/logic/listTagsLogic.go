package logic

import (
	"context"

	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTagsLogic {
	return &ListTagsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 标签定义列表：beid 必带，object_id<=0 时查全部对象，分页返回
func (l *ListTagsLogic) ListTags(in *tag.TagDefListReq) (*tag.TagDefListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, total, err := l.svcCtx.TagDefModel.FindList(l.ctx, in.Beid, in.ObjectId, in.Page, in.Size)
	if err != nil {
		l.Errorf("查询标签定义列表失败, beid: %d, object_id: %d, err: %v", in.Beid, in.ObjectId, err)
		return nil, err
	}
	items := make([]*tag.TagDef, 0, len(list))
	for _, d := range list {
		items = append(items, toPbTagDef(d))
	}
	return &tag.TagDefListReply{List: items, Total: total}, nil
}
