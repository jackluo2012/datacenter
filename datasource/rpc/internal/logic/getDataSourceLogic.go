package logic

import (
	"context"

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

// 数据源详情
func (l *GetDataSourceLogic) GetDataSource(in *datasource.GetDataSourceReq) (*datasource.DataSourceReply, error) {
	// todo: add your logic here and delete this line

	return &datasource.DataSourceReply{}, nil
}
