package logic

import (
	"context"

	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVersionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVersionsLogic {
	return &ListVersionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 版本列表
func (l *ListVersionsLogic) ListVersions(in *metadata.IdReq) (*metadata.VersionListReply, error) {
	// todo: add your logic here and delete this line

	return &metadata.VersionListReply{}, nil
}
