package logic

import (
	"context"

	"datacenter/datasource/model"
	"datacenter/datasource/rpc/datasource"
	"datacenter/datasource/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDataSourceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDataSourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDataSourceLogic {
	return &GetDataSourceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 数据源详情: with_secret=true 时解密密码并拼 DSN(仅内部服务调用), mysql 类型返回连接串
func (l *GetDataSourceLogic) GetDataSource(in *datasource.GetDataSourceReq) (*datasource.DataSourceReply, error) {
	// 1.查询(校验存在性与 beid 归属)
	ds, err := l.svcCtx.DataSourceModel.FindOne(in.Id, in.Beid)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, ErrNotFound
		}
		l.Errorf("查询数据源失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, ErrQueryFail
	}
	reply := &datasource.DataSourceReply{Datasource: toPbDataSource(*ds)}
	if !in.WithSecret {
		return reply, nil
	}
	// 2.with_secret: 解密密码(空密码返回空串), mysql 类型拼 DSN
	plain, err := DecryptPassword(l.svcCtx.Config.EncryptKey, ds.PasswordEnc)
	if err != nil {
		l.Errorf("解密密码失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, ErrDecryptFail
	}
	if ds.Type == "mysql" {
		reply.Dsn = BuildMysqlDsn(ds.Host, ds.Port, ds.DbName, ds.Username, plain)
	}
	return reply, nil
}
