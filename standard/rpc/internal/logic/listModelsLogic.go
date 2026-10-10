package logic

import (
	"context"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

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

// 标准模型列表（分页 + status 可选过滤 + keyword 对 model_code/model_name 模糊匹配）
func (l *ListModelsLogic) ListModels(in *standard.ModelListReq) (*standard.ModelListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.StdModelModel.FindList(l.ctx, in.Beid, in.Status, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	models := make([]*standard.StdModel, 0, len(list))
	for _, m := range list {
		models = append(models, modelToPb(m))
	}
	return &standard.ModelListReply{List: models, Total: total}, nil
}
