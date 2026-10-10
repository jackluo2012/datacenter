package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListModelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListModelsLogic {
	return &ListModelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListModelsLogic) ListModels(in *mdm.ModelListReq) (*mdm.ModelListReply, error) {
	// todo: add your logic here and delete this line

	return &mdm.ModelListReply{}, nil
}
