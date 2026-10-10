package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListModelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListModelsLogic {
	return &ListModelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListModels 分页查询主数据模型（keyword 模糊匹配 model_code/model_name）
func (l *ListModelsLogic) ListModels(in *mdm.ModelListReq) (*mdm.ModelListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, total, err := l.svcCtx.ModelModel.FindList(l.ctx, in.Beid, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	reply := &mdm.ModelListReply{Total: total}
	for _, row := range list {
		reply.List = append(reply.List, toPbModel(row))
	}
	return reply, nil
}
