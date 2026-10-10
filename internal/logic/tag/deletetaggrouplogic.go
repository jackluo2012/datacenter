// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTagGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除人群
func NewDeleteTagGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTagGroupLogic {
	return &DeleteTagGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTagGroupLogic) DeleteTagGroup(req *types.TagDefIdReq) error {
	// todo: add your logic here and delete this line

	return nil
}
