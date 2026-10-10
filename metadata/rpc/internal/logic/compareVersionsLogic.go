package logic

import (
	"context"

	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompareVersionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCompareVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompareVersionsLogic {
	return &CompareVersionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 版本对比
func (l *CompareVersionsLogic) CompareVersions(in *metadata.CompareReq) (*metadata.DiffReply, error) {
	// todo: add your logic here and delete this line

	return &metadata.DiffReply{}, nil
}
