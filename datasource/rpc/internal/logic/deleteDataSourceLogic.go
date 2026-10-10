package logic

import (
	"context"

	"datacenter/datasource/model"
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

// 删除数据源: 软删(仅写 deleted_at, 物理数据保留)
func (l *DeleteDataSourceLogic) DeleteDataSource(in *datasource.IdReq) (*datasource.OkReply, error) {
	// 1.先校验存在性与 beid 归属
	if _, err := l.svcCtx.DataSourceModel.FindOne(in.Id, in.Beid); err != nil {
		if err == model.ErrNotFound {
			return nil, ErrNotFound
		}
		l.Errorf("查询数据源失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, ErrQueryFail
	}
	// 2.软删
	if err := l.svcCtx.DataSourceModel.Delete(in.Id, in.Beid); err != nil {
		l.Errorf("删除数据源失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, ErrDeleteFail
	}
	return &datasource.OkReply{Ok: true}, nil
}
