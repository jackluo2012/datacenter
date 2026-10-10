// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package metadata

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/metadata/rpc/metadataClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type VersionCompareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 版本对比
func NewVersionCompareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VersionCompareLogic {
	return &VersionCompareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 版本对比: 网关请求逐字段映射为 rpc 请求后转发, 业务错误直接透传
func (l *VersionCompareLogic) VersionCompare(req *types.CompareReq) (resp *types.DiffResp, err error) {
	reply, err := l.svcCtx.MetadataRpc.CompareVersions(l.ctx, &metadataClient.CompareReq{
		VersionA: req.VersionA,
		VersionB: req.VersionB,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.DiffResp{
		DiffJson: reply.DiffJson,
	}, nil
}
