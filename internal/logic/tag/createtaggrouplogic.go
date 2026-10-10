// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTagGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建人群
func NewCreateTagGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTagGroupLogic {
	return &CreateTagGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTagGroupLogic) CreateTagGroup(req *types.CreateTagGroupReq) (resp *types.TagGroupInfo, err error) {
	// todo: add your logic here and delete this line

	return
}
