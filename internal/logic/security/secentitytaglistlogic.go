// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SecEntityTagListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSecEntityTagListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecEntityTagListLogic {
	return &SecEntityTagListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SecEntityTagListLogic) SecEntityTagList(req *types.EntityTagListReq) (resp *types.EntityTagListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
