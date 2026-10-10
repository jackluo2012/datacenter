package logic

import (
	"context"
	"database/sql"

	"datacenter/datasource/model"
	"datacenter/datasource/rpc/datasource"
	"datacenter/datasource/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDataSourceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDataSourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDataSourceLogic {
	return &UpdateDataSourceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新数据源: 字符串字段非空才覆盖, port 非零才覆盖, password 为空不更新密码列, status 以请求值为准(0.停用 1.启用)
func (l *UpdateDataSourceLogic) UpdateDataSource(in *datasource.UpdateDataSourceReq) (*datasource.OkReply, error) {
	// 1.查询原记录(校验存在性与 beid 归属)
	old, err := l.svcCtx.DataSourceModel.FindOne(in.Id, in.Beid)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, ErrNotFound
		}
		l.Errorf("查询数据源失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, ErrQueryFail
	}
	// 2.名称变更时做 name+beid 唯一性查重(排除自身)
	if in.Name != "" && in.Name != old.Name {
		dup, err := l.svcCtx.DataSourceModel.FindOneByName(in.Name, in.Beid)
		if err != nil && err != model.ErrNotFound {
			l.Errorf("查重失败, name: %s, beid: %d, err: %v", in.Name, in.Beid, err)
			return nil, ErrQueryFail
		}
		if dup != nil && dup.Id != in.Id {
			return nil, ErrNameDuplicate
		}
	}
	// 3.合并更新字段: 空值保留原值(proto3 无法区分未传, 按零值即不改处理)
	data := *old
	if in.Name != "" {
		data.Name = in.Name
	}
	if in.Type != "" {
		if !validDataSourceType(in.Type) {
			return nil, ErrTypeInvalid
		}
		data.Type = in.Type
	}
	if in.Host != "" {
		data.Host = in.Host
	}
	if in.Port != 0 {
		data.Port = in.Port
	}
	if in.DbName != "" {
		data.DbName = in.DbName
	}
	if in.Username != "" {
		data.Username = in.Username
	}
	if in.ParamsJson != "" {
		data.ParamsJson = sql.NullString{String: in.ParamsJson, Valid: true}
	}
	if in.Owner != "" {
		data.Owner = in.Owner
	}
	if err := l.svcCtx.DataSourceModel.Update(data); err != nil {
		l.Errorf("更新数据源失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, ErrUpdateFail
	}
	// 4.password 为空不更新密码列, 非空则重新加密
	if in.Password != "" {
		enc, err := EncryptPassword(l.svcCtx.Config.EncryptKey, in.Password)
		if err != nil {
			return nil, err
		}
		if err := l.svcCtx.DataSourceModel.UpdatePassword(in.Id, in.Beid, enc); err != nil {
			l.Errorf("更新密码失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
			return nil, ErrUpdateFail
		}
	}
	// 5.状态以请求值为准(0.停用 1.启用)
	if err := l.svcCtx.DataSourceModel.UpdateStatus(in.Id, in.Beid, in.Status); err != nil {
		l.Errorf("更新状态失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, ErrUpdateFail
	}
	return &datasource.OkReply{Ok: true}, nil
}
