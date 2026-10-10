package logic

import (
	"context"

	"datacenter/metadata/model"
	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListEntitiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListEntitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEntitiesLogic {
	return &ListEntitiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 实体列表
func (l *ListEntitiesLogic) ListEntities(in *metadata.EntityListReq) (*metadata.EntityListReply, error) {
	// 多租户条件必填
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.EntityModel.FindList(l.ctx, in.Beid, in.EntityType, in.Keyword, in.DsId, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	entities := make([]*metadata.MetaEntity, 0, len(list))
	for _, e := range list {
		entities = append(entities, entityToPb(e))
	}
	return &metadata.EntityListReply{List: entities, Total: total}, nil
}
