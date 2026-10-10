package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/metric/model"
	"datacenter/metric/rpc/internal/svc"
	"datacenter/metric/rpc/metric"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMetricLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateMetricLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMetricLogic {
	return &CreateMetricLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建指标：name_en/name_cn/metric_type 必填，按类型校验口径字段，name_en+beid 查重
func (l *CreateMetricLogic) CreateMetric(in *metric.CreateMetricReq) (*metric.MetricReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	nameEn := strings.TrimSpace(in.NameEn)
	if nameEn == "" {
		return nil, ErrNameEnRequired
	}
	if strings.TrimSpace(in.NameCn) == "" {
		return nil, ErrNameCnRequired
	}
	if in.MetricType == "" {
		return nil, ErrMetricTypeRequired
	}
	if !IsValidMetricType(in.MetricType) {
		return nil, ErrInvalidMetricType
	}
	// 按类型校验口径字段：atom 需 biz_field+agg_type（白名单），derive 需 expr
	if err := validateMetricSpec(in.MetricType, in.BizField, in.AggType, in.Expr); err != nil {
		return nil, err
	}

	// 英文名查重（uk_beid_name）
	_, err := l.svcCtx.MetricModel.FindOneByNameEn(l.ctx, in.Beid, nameEn)
	switch {
	case err == nil:
		return nil, ErrDuplicateMetric
	case !errors.Is(err, model.ErrNotFound):
		return nil, err
	}

	m := &model.Metric{
		NameEn:        nameEn,
		NameCn:        in.NameCn,
		MetricType:    in.MetricType,
		BizField:      in.BizField,
		AggType:       in.AggType,
		Expr:          in.Expr,
		DimensionJson: in.DimensionJson,
		Comment:       in.Comment,
		Owner:         in.Owner,
		Status:        1, // 默认启用
		Beid:          in.Beid,
	}
	id, err := l.svcCtx.MetricModel.Insert(l.ctx, m)
	if err != nil {
		return nil, err
	}
	m.Id = id

	l.Infof("[CreateMetric] id=%d name_en=%s type=%s beid=%d", id, m.NameEn, m.MetricType, m.Beid)
	return &metric.MetricReply{Metric: metricToPb(m)}, nil
}
