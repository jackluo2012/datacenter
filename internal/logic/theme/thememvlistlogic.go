// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/theme/rpc/themeClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ThemeMvListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 物化视图列表
func NewThemeMvListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThemeMvListLogic {
	return &ThemeMvListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 物化视图列表: 页码默认值处理后转发, 业务错误直接透传。
// 空列表仍 make 出空切片, 保证 JSON 序列化为 [] 而非 null。
func (l *ThemeMvListLogic) ThemeMvList(req *types.ThemeMvListReq) (resp *types.ThemeMvListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.ThemeRpc.ListMaterializedViews(l.ctx, &themeClient.MvListReq{
		Beid: req.Beid,
		Page: page,
		Size: size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.ThemeMvInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *mvToTypes(item))
	}
	return &types.ThemeMvListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
