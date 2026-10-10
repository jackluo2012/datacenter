package logic

import (
	"context"

	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApiLogic {
	return &GetApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// API详情
func (l *GetApiLogic) GetApi(in *datasvc.PathReq) (*datasvc.ApiReply, error) {
	// todo: add your logic here and delete this line

	return &datasvc.ApiReply{}, nil
}
