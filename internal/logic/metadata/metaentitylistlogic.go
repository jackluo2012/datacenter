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

type MetaEntityListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 实体列表
func NewMetaEntityListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MetaEntityListLogic {
	return &MetaEntityListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 实体列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *MetaEntityListLogic) MetaEntityList(req *types.MetaEntityListReq) (resp *types.MetaEntityListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.MetadataRpc.ListEntities(l.ctx, &metadataClient.EntityListReq{
		Beid:       req.Beid,
		EntityType: req.EntityType,
		Keyword:    req.Keyword,
		DsId:       req.DsId,
		Page:       page,
		Size:       size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.MetaEntity, 0, len(reply.List))
	for _, item := range reply.List {
		list = append(list, *entityToTypes(item))
	}
	return &types.MetaEntityListResp{
		List:  list,
		Total: reply.Total,
	}, nil
}
