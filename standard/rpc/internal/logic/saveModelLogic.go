package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveModelLogic {
	return &SaveModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存标准模型（id>0 更新，id=0 新建；elements_json 须为合法 JSON 数组
// 且引用的 element_code 全部存在；保存不动 status/version/快照）
func (l *SaveModelLogic) SaveModel(in *standard.SaveModelReq) (*standard.ModelReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	modelCode := strings.TrimSpace(in.ModelCode)
	if modelCode == "" {
		return nil, ErrModelCodeRequired
	}
	modelName := strings.TrimSpace(in.ModelName)
	if modelName == "" {
		return nil, ErrModelNameRequired
	}
	// elements_json 静态校验（JSON 数组、element_code 非空且不重复）
	elements, err := ParseElementsJson(in.ElementsJson)
	if err != nil {
		return nil, err
	}
	// 引用存在性校验：element_code 必须已在数据元表登记
	if _, err := resolveElements(l.ctx, l.svcCtx.ElementModel, in.Beid, elements); err != nil {
		return nil, err
	}

	data := &model.StdModel{
		ModelCode:    modelCode,
		ModelName:    modelName,
		ElementsJson: in.ElementsJson,
		Status:       model.ModelStatusDraft,
		Version:      1,
		Beid:         in.Beid,
	}

	if in.Id > 0 {
		// 更新：先确认存在，再对编码查重（排除自身）
		old, err := l.svcCtx.StdModelModel.FindOne(l.ctx, in.Id, in.Beid)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrModelNotFound
			}
			return nil, err
		}
		if dup, derr := l.svcCtx.StdModelModel.FindOneByCode(l.ctx, in.Beid, modelCode); derr == nil && dup.Id != old.Id {
			return nil, ErrModelCodeDuplicate
		} else if derr != nil && !errors.Is(derr, model.ErrNotFound) {
			return nil, derr
		}
		data.Id = old.Id
		if err := l.svcCtx.StdModelModel.Update(l.ctx, data); err != nil {
			return nil, err
		}
	} else {
		// 新建：编码查重后插入（draft 起步，版本 1）
		if _, err := l.svcCtx.StdModelModel.FindOneByCode(l.ctx, in.Beid, modelCode); err == nil {
			return nil, ErrModelCodeDuplicate
		} else if !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
		id, err := l.svcCtx.StdModelModel.Insert(l.ctx, data)
		if err != nil {
			return nil, err
		}
		data.Id = id
	}

	// 回读保证返回与库内一致
	m, err := l.svcCtx.StdModelModel.FindOne(l.ctx, data.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &standard.ModelReply{Model: modelToPb(m)}, nil
}

// resolveElements 按 elements_json 引用批量加载数据元定义，返回 code->数据元映射；
// 任一引用不存在（已删/跨租户）即报错，缺失编码体现在错误信息里
func resolveElements(ctx context.Context, m model.ElementModel, beid int64,
	elements []ModelElement) (map[string]*model.Element, error) {
	rows, err := m.FindListByCodes(ctx, beid, ElementCodes(elements))
	if err != nil {
		return nil, err
	}
	byCode := make(map[string]*model.Element, len(rows))
	for _, r := range rows {
		byCode[r.ElementCode] = r
	}
	var missing []string
	for _, e := range elements {
		if _, ok := byCode[e.ElementCode]; !ok {
			missing = append(missing, e.ElementCode)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: 缺失编码 %v", ErrElementMissing, missing)
	}
	return byCode, nil
}
