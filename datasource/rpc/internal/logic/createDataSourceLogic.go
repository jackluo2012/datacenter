package logic

import (
	"context"

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

// 创建数据源
func (l *CreateDataSourceLogic) CreateDataSource(in *datasource.CreateDataSourceReq) (*datasource.DataSourceReply, error) {
	// todo: add your logic here and delete this line

	return &datasource.DataSourceReply{}, nil
}
