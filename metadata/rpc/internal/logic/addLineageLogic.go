package logic

import (
	"context"

	"datacenter/metadata/model"
	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddLineageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddLineageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLineageLogic {
	return &AddLineageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登记血缘
func (l *AddLineageLogic) AddLineage(in *metadata.AddLineageReq) (*metadata.OkReply, error) {
	if in.FromId <= 0 || in.ToId <= 0 {
		return nil, ErrInvalidLineageParams
	}
	// rel_type=lineage，uk_edge(rel_type,from_id,to_id) 冲突忽略，重复登记幂等
	err := l.svcCtx.RelationModel.Insert(l.ctx, nil, &model.Relation{
		RelType:   model.RelTypeLineage,
		FromId:    in.FromId,
		ToId:      in.ToId,
		PropsJson: in.PropsJson,
		Beid:      in.Beid,
	})
	if err != nil {
		return nil, err
	}
	return &metadata.OkReply{Ok: true}, nil
}
