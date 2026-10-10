package logic

import (
	"context"

	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInfoLogic {
	return &GetInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登录增强
func (l *GetInfoLogic) GetInfo(in *sys.GetInfoReq) (*sys.GetInfoReply, error) {
	// todo: add your logic here and delete this line

	return &sys.GetInfoReply{}, nil
}
