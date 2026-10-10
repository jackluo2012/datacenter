package logic

import (
	"context"
	"errors"

	"datacenter/theme/model"
	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveThemeModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveThemeModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveThemeModelLogic {
	return &SaveThemeModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SaveThemeModel 新增/更新数仓分层模型：所属主题须存在；layer 白名单
// （ods/dwd/dws/ads/dim）；table_name 须通过小写标识符正则且 <layer>_<table> 整体
// 不超 MySQL 标识符上限（部署时按此真实建表，防注入）；fields_json 校验
// （name 防注入不重复 + type 白名单 + 主键恰有一个）；dim_json 可选校验；
// 同租户下 table_name 查重（uk_beid_table）
func (l *SaveThemeModelLogic) SaveThemeModel(in *theme.SaveThemeModelReq) (*theme.ModelReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	tableName := trimSpace(in.TableName)
	modelName := trimSpace(in.ModelName)
	if in.ThemeId <= 0 || tableName == "" || modelName == "" {
		return nil, ErrInvalidParams
	}
	//主题存在性校验（多租户过滤防跨租户挂靠）
	if _, err := l.svcCtx.ThemeModel.FindOne(l.ctx, in.ThemeId, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrThemeNotFound
		}
		return nil, err
	}
	//分层白名单
	layer := trimSpace(in.Layer)
	if !ValidLayer(layer) {
		return nil, ErrInvalidLayer
	}
	//表名防注入 + 物理表名（layer_table）长度把守
	if !ValidIdentifier(tableName) || len(tableName) > identifierMaxLen ||
		len(buildPhysicalTable(layer, tableName)) > identifierMaxLen {
		return nil, ErrInvalidTableName
	}
	//字段定义校验（name/type 白名单 + pk 恰有一个）并收敛为规范 JSON
	fields, err := ValidateFieldsJson(in.FieldsJson)
	if err != nil {
		return nil, err
	}
	fieldsJson, err := marshalFields(fields)
	if err != nil {
		return nil, err
	}
	//维度声明（可选）：非空须为 [{dim_table,join_key}]
	dimJson, err := ValidateDimJson(in.DimJson)
	if err != nil {
		return nil, err
	}
	//目标库名（可选）：非空须通过标识符校验（部署时拼进 DDL，防注入）
	dbName := trimSpace(in.DbName)
	if dbName != "" && (!ValidIdentifier(dbName) || len(dbName) > identifierMaxLen) {
		return nil, ErrInvalidDbName
	}

	//id=0 新建，id>0 更新
	if in.Id == 0 {
		//查重：同租户下 table_name 唯一（uk_beid_table）
		if _, err := l.svcCtx.ModelModel.FindByTable(l.ctx, in.Beid, tableName); err == nil {
			return nil, ErrDupTableName
		} else if !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
		row := &model.ThemeModelRow{
			ThemeId:      in.ThemeId,
			Layer:        layer,
			TableName:    tableName,
			ModelName:    modelName,
			FieldsJson:   fieldsJson,
			DimJson:      dimJson,
			DatasourceId: in.DatasourceId,
			DbName:       dbName,
			Beid:         in.Beid,
		}
		id, err := l.svcCtx.ModelModel.Insert(l.ctx, row)
		if err != nil {
			if errors.Is(err, model.ErrDupTableName) {
				return nil, ErrDupTableName
			}
			return nil, err
		}
		row.Id = id
		l.Infof("[SaveThemeModel] 新建模型 id=%d table=%s layer=%s beid=%d", id, tableName, layer, in.Beid)
		return &theme.ModelReply{Model: toPbThemeModel(row)}, nil
	}

	//更新：须存在
	exist, err := l.svcCtx.ModelModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}
	//表名若变更，仍须查重
	if exist.TableName != tableName {
		if dup, err := l.svcCtx.ModelModel.FindByTable(l.ctx, in.Beid, tableName); err == nil && dup.Id != exist.Id {
			return nil, ErrDupTableName
		} else if err != nil && !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
	}
	row := &model.ThemeModelRow{
		Id:           exist.Id,
		ThemeId:      in.ThemeId,
		Layer:        layer,
		TableName:    tableName,
		ModelName:    modelName,
		FieldsJson:   fieldsJson,
		DimJson:      dimJson,
		DatasourceId: in.DatasourceId,
		DbName:       dbName,
		Beid:         in.Beid,
	}
	if err := l.svcCtx.ModelModel.Update(l.ctx, row); err != nil {
		if errors.Is(err, model.ErrDupTableName) {
			return nil, ErrDupTableName
		}
		return nil, err
	}
	l.Infof("[SaveThemeModel] 更新模型 id=%d table=%s layer=%s beid=%d", row.Id, tableName, layer, in.Beid)
	return &theme.ModelReply{Model: toPbThemeModel(row)}, nil
}
