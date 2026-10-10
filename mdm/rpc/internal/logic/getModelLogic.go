package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetModelLogic {
	return &GetModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetModelLogic) GetModel(in *mdm.IdReq) (*mdm.ModelReply, error) {
	// todo: add your logic here and delete this line

	return &mdm.ModelReply{}, nil
}
