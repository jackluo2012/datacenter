package logic

import (
	"context"

	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAtomsOfDeriveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAtomsOfDeriveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAtomsOfDeriveLogic {
	return &ListAtomsOfDeriveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取派生指标引用的原子指标
func (l *ListAtomsOfDeriveLogic) ListAtomsOfDerive(in *metric.IdReq) (*metric.MetricListReply, error) {
	// todo: add your logic here and delete this line

	return &metric.MetricListReply{}, nil
}
