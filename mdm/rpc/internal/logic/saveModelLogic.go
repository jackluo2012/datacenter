package logic

import (
	"context"
	"errors"

	"datacenter/mdm/model"
	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

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

// SaveModel 新增/更新主数据模型：model_code/name 必填查重；fields_json 须为 JSON 数组
// （每项含 name/type，type 白名单）；code_strategy 仅 segment/auto；table_name 为空按
// `dc_mdm_data_<model_code>` 自动生成且必须通过小写标识符校验（物理表会真实建表，防注入）
func (l *SaveModelLogic) SaveModel(in *mdm.SaveModelReq) (*mdm.ModelReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	modelCode := trimSpace(in.ModelCode)
	modelName := trimSpace(in.ModelName)
	if modelCode == "" || modelName == "" {
		return nil, ErrInvalidParams
	}

	//字段定义校验（name/type 必备、type 白名单、name 防注入/防内置列冲突/防重复）
	fields, err := ValidateFieldsJson(in.FieldsJson)
	if err != nil {
		return nil, err
	}
	fieldsJson, err := marshalFields(fields)
	if err != nil {
		return nil, err
	}

	//编码策略白名单（空缺省 segment）
	strategy := trimSpace(in.CodeStrategy)
	if strategy == "" {
		strategy = CodeStrategySegment
	}
	if !ValidCodeStrategy(strategy) {
		return nil, ErrInvalidCodeStrategy
	}

	//表名规范化：为空自动生成，非法即拒绝
	tableName, err := NormalizeTableName(modelCode, in.TableName)
	if err != nil {
		return nil, err
	}

	//id=0 新建，id>0 更新
	if in.Id == 0 {
		//查重：同租户下 model_code 唯一
		if _, err := l.svcCtx.ModelModel.FindByCode(l.ctx, in.Beid, modelCode); err == nil {
			return nil, ErrDupModelCode
		} else if !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
		row := &model.MdmModelRow{
			ModelCode:    modelCode,
			ModelName:    modelName,
			TableName:    tableName,
			FieldsJson:   fieldsJson,
			CodeStrategy: strategy,
			Beid:         in.Beid,
		}
		id, err := l.svcCtx.ModelModel.Insert(l.ctx, row)
		if err != nil {
			if errors.Is(err, model.ErrDupModelCode) {
				return nil, ErrDupModelCode
			}
			return nil, err
		}
		row.Id = id
		l.Infof("[SaveModel] 新建模型 id=%d code=%s table=%s beid=%d", id, modelCode, tableName, in.Beid)
		return &mdm.ModelReply{Model: toPbModel(row)}, nil
	}

	//更新：须存在
	exist, err := l.svcCtx.ModelModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}
	//编码若变更，仍须查重
	if exist.ModelCode != modelCode {
		if dup, err := l.svcCtx.ModelModel.FindByCode(l.ctx, in.Beid, modelCode); err == nil && dup.Id != exist.Id {
			return nil, ErrDupModelCode
		} else if err != nil && !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
	}
	row := &model.MdmModelRow{
		Id:           exist.Id,
		ModelCode:    modelCode,
		ModelName:    modelName,
		TableName:    tableName,
		FieldsJson:   fieldsJson,
		CodeStrategy: strategy,
		Beid:         in.Beid,
	}
	if err := l.svcCtx.ModelModel.Update(l.ctx, row); err != nil {
		if errors.Is(err, model.ErrDupModelCode) {
			return nil, ErrDupModelCode
		}
		return nil, err
	}
	l.Infof("[SaveModel] 更新模型 id=%d code=%s table=%s beid=%d", row.Id, modelCode, tableName, in.Beid)
	return &mdm.ModelReply{Model: toPbModel(row)}, nil
}
