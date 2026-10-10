package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteGrantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteGrantLogic {
	return &DeleteGrantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteGrantLogic) DeleteGrant(in *security.IdReq) (*security.IdReq, error) {
	// todo: add your logic here and delete this line

	return &security.IdReq{}, nil
}
