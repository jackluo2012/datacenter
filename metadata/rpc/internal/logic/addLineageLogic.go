package logic

import (
	"context"

	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddLineageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddLineageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLineageLogic {
	return &AddLineageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登记血缘
func (l *AddLineageLogic) AddLineage(in *metadata.AddLineageReq) (*metadata.OkReply, error) {
	// todo: add your logic here and delete this line

	return &metadata.OkReply{}, nil
}
