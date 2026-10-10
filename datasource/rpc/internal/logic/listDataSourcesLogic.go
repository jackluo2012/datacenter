package logic

import (
	"context"

	"datacenter/datasource/rpc/datasource"
	"datacenter/datasource/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// 分页参数上限
const (
	defaultPageSize = 10
	maxPageSize     = 100
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

// 数据源列表: beid 必带, type 可选过滤, 分页返回
func (l *ListDataSourcesLogic) ListDataSources(in *datasource.ListReq) (*datasource.DataSourceListReply, error) {
	// 分页参数兜底
	page, size := in.Page, in.Size
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	list, total, err := l.svcCtx.DataSourceModel.FindList(in.Beid, in.Type, page, size)
	if err != nil {
		l.Errorf("查询数据源列表失败, beid: %d, type: %s, err: %v", in.Beid, in.Type, err)
		return nil, ErrQueryFail
	}
	resp := &datasource.DataSourceListReply{Total: total}
	for i := range list {
		resp.List = append(resp.List, toPbDataSource(list[i]))
	}
	return resp, nil
}
