package logic

import (
	"context"

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

// 更新数据源
func (l *UpdateDataSourceLogic) UpdateDataSource(in *datasource.UpdateDataSourceReq) (*datasource.OkReply, error) {
	// todo: add your logic here and delete this line

	return &datasource.OkReply{}, nil
}
