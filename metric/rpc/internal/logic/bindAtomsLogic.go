package logic

import (
	"context"

	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindAtomsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindAtomsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindAtomsLogic {
	return &BindAtomsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 派生指标绑定原子指标（幂等覆盖）
func (l *BindAtomsLogic) BindAtoms(in *metric.BindAtomReq) (*metric.IdReq, error) {
	// todo: add your logic here and delete this line

	return &metric.IdReq{}, nil
}
