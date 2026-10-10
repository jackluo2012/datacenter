package logic

import (
	"context"

	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoutersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRoutersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoutersLogic {
	return &GetRoutersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRoutersLogic) GetRouters(in *sys.GetRoutersReq) (*sys.GetRoutersReply, error) {
	// todo: add your logic here and delete this line

	return &sys.GetRoutersReply{}, nil
}
