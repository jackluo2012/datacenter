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
	list, err := l.svcCtx.VersionModel.FindList(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	versions := make([]*metadata.MetaVersion, 0, len(list))
	for _, v := range list {
		versions = append(versions, versionToPb(v))
	}
	return &metadata.VersionListReply{List: versions}, nil
}
