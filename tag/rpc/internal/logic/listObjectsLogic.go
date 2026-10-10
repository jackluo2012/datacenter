package logic

import (
	"context"

	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListObjectsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListObjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListObjectsLogic {
	return &ListObjectsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 对象列表：beid 必带，object_type 可选过滤，分页返回
func (l *ListObjectsLogic) ListObjects(in *tag.ObjectListReq) (*tag.ObjectListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, total, err := l.svcCtx.ObjectModel.FindList(l.ctx, in.Beid, in.ObjectType, in.Page, in.Size)
	if err != nil {
		l.Errorf("查询标签对象列表失败, beid: %d, err: %v", in.Beid, err)
		return nil, err
	}
	items := make([]*tag.TagObject, 0, len(list))
	for _, o := range list {
		items = append(items, toPbObject(o))
	}
	return &tag.ObjectListReply{List: items, Total: total}, nil
}
