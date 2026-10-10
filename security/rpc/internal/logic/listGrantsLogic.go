package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListGrantsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListGrantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListGrantsLogic {
	return &ListGrantsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListGrantsLogic) ListGrants(in *security.GrantListReq) (*security.GrantListReply, error) {
	// todo: add your logic here and delete this line

	return &security.GrantListReply{}, nil
}
