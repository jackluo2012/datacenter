package logic

import (
	"context"

	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type CollectDatabaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCollectDatabaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectDatabaseLogic {
	return &CollectDatabaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 采集库表元数据（含 contains 关系）
func (l *CollectDatabaseLogic) CollectDatabase(in *metadata.CollectDatabaseReq) (*metadata.CollectReply, error) {
	// todo: add your logic here and delete this line

	return &metadata.CollectReply{}, nil
}
