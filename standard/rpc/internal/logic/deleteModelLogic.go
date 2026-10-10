package logic

import (
	"context"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteModelLogic {
	return &DeleteModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除标准模型（幂等：模型不存在时不报错；历史核对报告按设计保留不级联删）
func (l *DeleteModelLogic) DeleteModel(in *standard.IdReq) (*standard.IdReq, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	if err := l.svcCtx.StdModelModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	return &standard.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
