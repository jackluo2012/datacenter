// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/tag/rpc/tagClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type TagGroupMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 人群成员
func NewTagGroupMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagGroupMembersLogic {
	return &TagGroupMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TagGroupMembersLogic) TagGroupMembers(req *types.GroupMembersReq) (resp *types.GroupMembersResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.TagRpc.ListGroupMembers(l.ctx, &tagClient.GroupMembersReq{
		GroupId: req.GroupId,
		Beid:    req.Beid,
		Page:    int64(page),
		Size:    int64(size),
	})
	if err != nil {
		return nil, err
	}
	members := make([]types.GroupMember, 0, len(reply.GetMembers()))
	for _, m := range reply.GetMembers() {
		members = append(members, *memberToTypes(m))
	}
	return &types.GroupMembersResp{Members: members, Total: reply.GetTotal()}, nil
}
