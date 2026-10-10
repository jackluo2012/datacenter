// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package metric

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/metric/rpc/metricsvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMetricLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新指标
func NewUpdateMetricLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMetricLogic {
	return &UpdateMetricLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 更新指标: 网关请求逐字段映射为 rpc 请求后转发; pb UpdateMetric 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *UpdateMetricLogic) UpdateMetric(req *types.UpdateMetricReq) error {
	_, err := l.svcCtx.MetricRpc.UpdateMetric(l.ctx, &metricsvc.UpdateMetricReq{
		Id:            req.Id,
		Beid:          req.Beid,
		NameCn:        req.NameCn,
		Expr:          req.Expr,
		DimensionJson: req.DimensionJson,
		Comment:       req.Comment,
		Status:        req.Status,
	})
	return err
}
