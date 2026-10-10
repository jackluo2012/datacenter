package logic

import (
	"context"

	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPwdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetPwdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPwdLogic {
	return &ResetPwdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResetPwdLogic) ResetPwd(in *sys.ResetPwdReq) (*sys.IdReq, error) {
	// todo: add your logic here and delete this line

	return &sys.IdReq{}, nil
}
