package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishModelLogic {
	return &PublishModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// snapshotDoc 发布快照文档：模型定义 + 组成元素的完整定义冻结（CheckTable 按此核对）
type snapshotDoc struct {
	ModelCode   string            `json:"model_code"`
	ModelName   string            `json:"model_name"`
	Version     int32             `json:"version"`
	PublishedAt int64             `json:"published_at"` //毫秒时间戳
	Elements    []snapshotElement `json:"elements"`
}

// snapshotElement 快照内的数据元定义（冗余元素全量属性，发布后不随数据元编辑漂移）
type snapshotElement struct {
	ElementCode   string `json:"element_code"`
	ElementName   string `json:"element_name"`
	DataType      string `json:"data_type"`
	DataLength    int32  `json:"data_length"`
	DataPrecision int32  `json:"data_precision"`
	DictCode      string `json:"dict_code,omitempty"`
	Required      bool   `json:"required"`
	Pk            bool   `json:"pk"`
}

// 发布标准模型：校验 elements 完整（非空 + 每个数据元存在）→ version+1 →
// 组装快照（模型 + 数据元定义 JSON）落 snapshot_json，status 置 published
func (l *PublishModelLogic) PublishModel(in *standard.PublishReq) (*standard.PublishReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	m, err := l.svcCtx.StdModelModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}
	// elements 完整性校验：非空且引用的数据元全部存在
	elements, err := ParseElementsJson(m.ElementsJson)
	if err != nil {
		return nil, err
	}
	if len(elements) == 0 {
		return nil, ErrElementsJsonInvalid
	}
	byCode, err := resolveElements(l.ctx, l.svcCtx.ElementModel, in.Beid, elements)
	if err != nil {
		return nil, err
	}

	// 组装快照：模型 + 数据元定义一次性冻结
	publishedAt := time.Now()
	snapshot := snapshotDoc{
		ModelCode:   m.ModelCode,
		ModelName:   m.ModelName,
		Version:     m.Version + 1,
		PublishedAt: publishedAt.UnixMilli(),
		Elements:    make([]snapshotElement, 0, len(elements)),
	}
	for _, e := range elements {
		def := byCode[e.ElementCode]
		snapshot.Elements = append(snapshot.Elements, snapshotElement{
			ElementCode:   def.ElementCode,
			ElementName:   def.ElementName,
			DataType:      def.DataType,
			DataLength:    def.DataLength,
			DataPrecision: def.DataPrecision,
			DictCode:      def.DictCode,
			Required:      e.Required,
			Pk:            e.Pk,
		})
	}
	snapshotJson, err := json.Marshal(&snapshot)
	if err != nil {
		return nil, err
	}

	// version+1 与快照一次落库，status 置 published
	newVersion := m.Version + 1
	if err := l.svcCtx.StdModelModel.Publish(l.ctx, m.Id, in.Beid, newVersion, string(snapshotJson), publishedAt); err != nil {
		return nil, err
	}
	l.Infof("标准模型已发布: id=%d code=%s version=%d operator=%s", m.Id, m.ModelCode, newVersion, in.Operator)
	return &standard.PublishReply{
		Version:      newVersion,
		SnapshotJson: string(snapshotJson),
	}, nil
}
