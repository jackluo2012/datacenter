// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
