package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveGrantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveGrantLogic {
	return &SaveGrantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 授权矩阵
func (l *SaveGrantLogic) SaveGrant(in *security.SaveGrantReq) (*security.GrantReply, error) {
	// todo: add your logic here and delete this line

	return &security.GrantReply{}, nil
}
