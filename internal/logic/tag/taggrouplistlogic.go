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

type TagGroupListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 人群列表
func NewTagGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagGroupListLogic {
	return &TagGroupListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TagGroupListLogic) TagGroupList(req *types.TagGroupListReq) (resp *types.TagGroupListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.TagRpc.ListGroups(l.ctx, &tagClient.GroupListReq{
		Beid: req.Beid,
		Page: int64(page),
		Size: int64(size),
	})
	if err != nil {
		return nil, err
	}
	return &types.TagGroupListResp{List: groupListToTypes(reply.GetList()), Total: reply.GetTotal()}, nil
}
