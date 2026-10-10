package logic

import (
	"context"

	"datacenter/datasource/rpc/datasource"
	"datacenter/datasource/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDataSourcesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDataSourcesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDataSourcesLogic {
	return &ListDataSourcesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 数据源列表
func (l *ListDataSourcesLogic) ListDataSources(in *datasource.ListReq) (*datasource.DataSourceListReply, error) {
	// todo: add your logic here and delete this line

	return &datasource.DataSourceListReply{}, nil
}
