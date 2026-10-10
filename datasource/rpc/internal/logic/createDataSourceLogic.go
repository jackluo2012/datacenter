package logic

import (
	"context"
	"database/sql"

	"datacenter/datasource/model"
	"datacenter/datasource/rpc/datasource"
	"datacenter/datasource/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDataSourceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateDataSourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDataSourceLogic {
	return &CreateDataSourceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建数据源: 校验必填 -> name+beid 查重 -> 密码 AES-GCM 加密 -> 落库 -> 返回(不含敏感信息)
func (l *CreateDataSourceLogic) CreateDataSource(in *datasource.CreateDataSourceReq) (*datasource.DataSourceReply, error) {
	// 1.校验必填字段与类型白名单
	if in.Name == "" {
		return nil, ErrNameEmpty
	}
	if !validDataSourceType(in.Type) {
		return nil, ErrTypeInvalid
	}
	// 2.name+beid 唯一性查重(未删除范围内)
	if _, err := l.svcCtx.DataSourceModel.FindOneByName(in.Name, in.Beid); err == nil {
		return nil, ErrNameDuplicate
	} else if err != model.ErrNotFound {
		l.Errorf("查重失败, name: %s, beid: %d, err: %v", in.Name, in.Beid, err)
		return nil, ErrQueryFail
	}
	// 3.明文密码 AES-GCM 加密(空密码加密为 nonce+空密文, 统一走同一逻辑)
	enc, err := EncryptPassword(l.svcCtx.Config.EncryptKey, in.Password)
	if err != nil {
		return nil, err
	}
	// 4.落库, 新数据源默认启用
	ret, err := l.svcCtx.DataSourceModel.Insert(model.DataSource{
		Name:        in.Name,
		Type:        in.Type,
		Host:        in.Host,
		Port:        in.Port,
		DbName:      in.DbName,
		Username:    in.Username,
		PasswordEnc: enc,
		ParamsJson:  sql.NullString{String: in.ParamsJson, Valid: in.ParamsJson != ""},
		Status:      1,
		Owner:       in.Owner,
		Beid:        in.Beid,
	})
	if err != nil {
		l.Errorf("创建数据源失败, name: %s, beid: %d, err: %v", in.Name, in.Beid, err)
		return nil, ErrCreateFail
	}
	id, err := ret.LastInsertId()
	if err != nil {
		l.Errorf("获取自增id失败, name: %s, beid: %d, err: %v", in.Name, in.Beid, err)
		return nil, ErrCreateFail
	}
	// 5.返回结果不回传任何敏感信息(密文也不回传)
	return &datasource.DataSourceReply{
		Datasource: &datasource.DataSourceInfo{
			Id:         id,
			Name:       in.Name,
			Type:       in.Type,
			Host:       in.Host,
			Port:       in.Port,
			DbName:     in.DbName,
			Username:   in.Username,
			ParamsJson: in.ParamsJson,
			Status:     1,
			Owner:      in.Owner,
			Beid:       in.Beid,
		},
	}, nil
}
