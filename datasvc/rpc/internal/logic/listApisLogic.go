package logic

import (
	"context"

	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListApisLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListApisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListApisLogic {
	return &ListApisLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// API列表
func (l *ListApisLogic) ListApis(in *datasvc.ApiListReq) (*datasvc.ApiListReply, error) {
	// todo: add your logic here and delete this line

	return &datasvc.ApiListReply{}, nil
}
