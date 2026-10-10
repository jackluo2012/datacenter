package logic

import (
	"context"

	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListGroupMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListGroupMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListGroupMembersLogic {
	return &ListGroupMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 人群成员查询
func (l *ListGroupMembersLogic) ListGroupMembers(in *tag.GroupMembersReq) (*tag.GroupMembersReply, error) {
	// todo: add your logic here and delete this line

	return &tag.GroupMembersReply{}, nil
}
