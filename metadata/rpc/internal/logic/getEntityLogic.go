package logic

import (
	"context"
	"errors"

	"datacenter/metadata/model"
	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEntityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEntityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEntityLogic {
	return &GetEntityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 实体详情
func (l *GetEntityLogic) GetEntity(in *metadata.IdReq) (*metadata.EntityReply, error) {
	if in.Id <= 0 {
		return nil, ErrEntityNotFound
	}
	entity, err := l.svcCtx.EntityModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}
	return &metadata.EntityReply{Entity: entityToPb(entity)}, nil
}
