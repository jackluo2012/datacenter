// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSecGrantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteSecGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSecGrantLogic {
	return &DeleteSecGrantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSecGrantLogic) DeleteSecGrant(req *types.SecIdReq) error {
	// todo: add your logic here and delete this line

	return nil
}
