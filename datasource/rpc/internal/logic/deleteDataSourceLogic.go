package logic

import (
	"context"

	"datacenter/datasource/rpc/datasource"
	"datacenter/datasource/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDataSourceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDataSourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDataSourceLogic {
	return &DeleteDataSourceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除数据源
func (l *DeleteDataSourceLogic) DeleteDataSource(in *datasource.IdReq) (*datasource.OkReply, error) {
	// todo: add your logic here and delete this line

	return &datasource.OkReply{}, nil
}
