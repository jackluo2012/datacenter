package logic

import (
	"context"
	"errors"

	"datacenter/metric/model"
	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMetricLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMetricLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMetricLogic {
	return &UpdateMetricLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新指标：部分更新，字符串字段（name_cn/expr/dimension_json/comment）非空才覆盖，
// status 约定仅非 0 时更新（proto3 无法区分未传与 0）
func (l *UpdateMetricLogic) UpdateMetric(in *metric.UpdateMetricReq) (*metric.IdReq, error) {
	if in.Id <= 0 {
		return nil, ErrMetricNotFound
	}
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	// 存在性校验（同时限定多租户）
	if _, err := l.svcCtx.MetricModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrMetricNotFound
		}
		return nil, err
	}

	// 字符串字段部分更新（model 内非空才拼 set）
	if err := l.svcCtx.MetricModel.Update(l.ctx, &model.Metric{
		Id:            in.Id,
		Beid:          in.Beid,
		NameCn:        in.NameCn,
		Expr:          in.Expr,
		DimensionJson: in.DimensionJson,
		Comment:       in.Comment,
	}); err != nil {
		return nil, err
	}

	// 状态单独更新（仅非 0 视为需要变更）
	if in.Status != 0 {
		if err := l.svcCtx.MetricModel.UpdateStatus(l.ctx, in.Id, in.Beid, int64(in.Status)); err != nil {
			return nil, err
		}
	}

	l.Infof("[UpdateMetric] id=%d beid=%d status=%d", in.Id, in.Beid, in.Status)
	return &metric.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
