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

type ThemeModelListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 数仓模型列表
func NewThemeModelListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThemeModelListLogic {
	return &ThemeModelListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 数仓模型列表: 页码默认值处理后转发(theme_id/layer 为过滤条件, 0/空串表示不过滤), 业务错误直接透传。
// 空列表仍 make 出空切片, 保证 JSON 序列化为 [] 而非 null。
func (l *ThemeModelListLogic) ThemeModelList(req *types.ThemeModelListReq) (resp *types.ThemeModelListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.ThemeRpc.ListThemeModels(l.ctx, &themeClient.ModelListReq{
		Beid:    req.Beid,
		ThemeId: req.ThemeId,
		Layer:   req.Layer,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.ThemeModelInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *themeModelToTypes(item))
	}
	return &types.ThemeModelListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
