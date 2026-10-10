package logic

import (
	"context"

	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEntityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEntityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEntityLogic {
	return &GetEntityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 实体详情
func (l *GetEntityLogic) GetEntity(in *metadata.IdReq) (*metadata.EntityReply, error) {
	// todo: add your logic here and delete this line

	return &metadata.EntityReply{}, nil
}
