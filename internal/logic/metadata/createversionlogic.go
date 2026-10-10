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

type CreateVersionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 定版快照
func NewCreateVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVersionLogic {
	return &CreateVersionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 定版快照: 网关请求逐字段映射为 rpc 请求后转发, 业务错误直接透传。
// 网关与 rpc 的请求结构均无 operator 字段, 操作人兜底 "admin" 由 rpc 侧负责。
func (l *CreateVersionLogic) CreateVersion(req *types.MetaEntityIdReq) (resp *types.VersionResp, err error) {
	reply, err := l.svcCtx.MetadataRpc.CreateVersion(l.ctx, &metadataClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.VersionResp{
		Version: *versionToTypes(reply.Version),
	}, nil
}
