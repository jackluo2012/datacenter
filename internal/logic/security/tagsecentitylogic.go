// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TagSecEntityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTagSecEntityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagSecEntityLogic {
	return &TagSecEntityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TagSecEntityLogic) TagSecEntity(req *types.TagSecEntityReq) (resp *types.SecEntityTag, err error) {
	// todo: add your logic here and delete this line

	return
}
