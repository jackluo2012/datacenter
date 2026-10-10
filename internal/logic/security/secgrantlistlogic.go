// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SecGrantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSecGrantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecGrantListLogic {
	return &SecGrantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SecGrantListLogic) SecGrantList(req *types.SecGrantListReq) (resp *types.SecGrantListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
