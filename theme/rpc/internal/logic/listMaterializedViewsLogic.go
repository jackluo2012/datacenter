package logic

import (
	"context"

	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMaterializedViewsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMaterializedViewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMaterializedViewsLogic {
	return &ListMaterializedViewsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListMaterializedViews 物化视图分页列表
func (l *ListMaterializedViewsLogic) ListMaterializedViews(in *theme.MvListReq) (*theme.MvListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, total, err := l.svcCtx.ThemeMvModel.FindList(l.ctx, in.Beid, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	pbList := make([]*theme.MaterializedView, 0, len(list))
	for _, m := range list {
		pbList = append(pbList, toPbMv(m))
	}
	return &theme.MvListReply{List: pbList, Total: total}, nil
}
