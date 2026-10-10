package logic

import (
	"context"

	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListGroupsLogic {
	return &ListGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 人群列表：beid 必带，分页返回（含 est_count/物化表名/最近圈选时间等运行状态）
func (l *ListGroupsLogic) ListGroups(in *tag.GroupListReq) (*tag.GroupListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, total, err := model.NewGroupModel(l.svcCtx.Conn).FindList(l.ctx, in.Beid, in.Page, in.Size)
	if err != nil {
		l.Errorf("查询人群列表失败, beid: %d, err: %v", in.Beid, err)
		return nil, err
	}
	items := make([]*tag.TagGroup, 0, len(list))
	for _, g := range list {
		items = append(items, toPbGroup(g))
	}
	return &tag.GroupListReply{List: items, Total: total}, nil
}
