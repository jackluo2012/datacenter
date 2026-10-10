package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveElementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveElementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveElementLogic {
	return &SaveElementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存数据元（id>0 更新，id=0 新建；element_code 同租户内查重；
// data_type 白名单校验，dict_code 非空时校验字典存在）
func (l *SaveElementLogic) SaveElement(in *standard.SaveElementReq) (*standard.ElementReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	elementCode := strings.TrimSpace(in.ElementCode)
	if elementCode == "" {
		return nil, ErrElementCodeRequired
	}
	elementName := strings.TrimSpace(in.ElementName)
	if elementName == "" {
		return nil, ErrElementNameRequired
	}
	// data_type 白名单 + 按类型的长度必填校验
	dataType := strings.TrimSpace(in.DataType)
	if !IsValidDataType(dataType) {
		return nil, ErrDataTypeInvalid
	}
	if err := validateDataTypeLength(dataType, in.DataLength); err != nil {
		return nil, err
	}
	// 值域引用：dict_code 非空时字典必须存在
	dictCode := strings.TrimSpace(in.DictCode)
	if dictCode != "" {
		if _, err := l.svcCtx.DictModel.FindOneByCode(l.ctx, nil, in.Beid, dictCode); err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrDictRefNotFound
			}
			return nil, err
		}
	}

	data := &model.Element{
		ElementCode:   elementCode,
		ElementName:   elementName,
		DataType:      dataType,
		DataLength:    in.DataLength,
		DataPrecision: in.DataPrecision,
		DictCode:      dictCode,
		Comment:       in.Comment,
		Beid:          in.Beid,
	}

	if in.Id > 0 {
		// 更新：先确认存在，再对编码查重（排除自身）
		old, err := l.svcCtx.ElementModel.FindOne(l.ctx, in.Id, in.Beid)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrElementNotFound
			}
			return nil, err
		}
		if dup, derr := l.svcCtx.ElementModel.FindOneByCode(l.ctx, in.Beid, elementCode); derr == nil && dup.Id != old.Id {
			return nil, ErrElementCodeDuplicate
		} else if derr != nil && !errors.Is(derr, model.ErrNotFound) {
			return nil, derr
		}
		data.Id = old.Id
		if err := l.svcCtx.ElementModel.Update(l.ctx, data); err != nil {
			return nil, err
		}
	} else {
		// 新建：编码查重后插入
		if _, err := l.svcCtx.ElementModel.FindOneByCode(l.ctx, in.Beid, elementCode); err == nil {
			return nil, ErrElementCodeDuplicate
		} else if !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
		id, err := l.svcCtx.ElementModel.Insert(l.ctx, data)
		if err != nil {
			return nil, err
		}
		data.Id = id
	}

	// 回读保证返回与库内一致
	element, err := l.svcCtx.ElementModel.FindOne(l.ctx, data.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &standard.ElementReply{Element: elementToPb(element)}, nil
}

// validateDataTypeLength 按类型校验长度（纯函数）：
// varchar/decimal 需要 data_length>0（varchar 为字符长度，decimal 为总精度）；
// int/bigint 的显示宽度 MySQL 8 已废弃不采集，date/datetime/text 无长度概念，均不强校验
func validateDataTypeLength(dataType string, length int32) error {
	switch dataType {
	case DataTypeVarchar, DataTypeDecimal:
		if length <= 0 {
			return ErrDataTypeLengthInvalid
		}
	}
	return nil
}
