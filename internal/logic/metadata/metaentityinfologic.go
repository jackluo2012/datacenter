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

type MetaEntityInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 实体详情
func NewMetaEntityInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MetaEntityInfoLogic {
	return &MetaEntityInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 实体详情: 网关请求逐字段映射为 rpc 请求后转发, 业务错误直接透传
func (l *MetaEntityInfoLogic) MetaEntityInfo(req *types.MetaEntityIdReq) (resp *types.MetaEntity, err error) {
	reply, err := l.svcCtx.MetadataRpc.GetEntity(l.ctx, &metadataClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return entityToTypes(reply.Entity), nil
}
