package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPermLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckPermLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPermLogic {
	return &CheckPermLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 权限检查
func (l *CheckPermLogic) CheckPerm(in *security.CheckPermReq) (*security.CheckPermReply, error) {
	// todo: add your logic here and delete this line

	return &security.CheckPermReply{}, nil
}
