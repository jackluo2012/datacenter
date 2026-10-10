package logic

import (
	"context"

	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLineageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLineageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLineageLogic {
	return &GetLineageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 血缘图（上下游 N 跳）
func (l *GetLineageLogic) GetLineage(in *metadata.LineageReq) (*metadata.LineageReply, error) {
	// todo: add your logic here and delete this line

	return &metadata.LineageReply{}, nil
}
