package logic

import (
	"context"

	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListEntitiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListEntitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEntitiesLogic {
	return &ListEntitiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 实体列表
func (l *ListEntitiesLogic) ListEntities(in *metadata.EntityListReq) (*metadata.EntityListReply, error) {
	// todo: add your logic here and delete this line

	return &metadata.EntityListReply{}, nil
}
