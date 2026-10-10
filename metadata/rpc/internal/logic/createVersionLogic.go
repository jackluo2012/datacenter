package logic

import (
	"context"

	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVersionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVersionLogic {
	return &CreateVersionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 定版快照
func (l *CreateVersionLogic) CreateVersion(in *metadata.IdReq) (*metadata.VersionReply, error) {
	// todo: add your logic here and delete this line

	return &metadata.VersionReply{}, nil
}
