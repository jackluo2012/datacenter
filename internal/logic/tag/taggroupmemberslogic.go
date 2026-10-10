// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
