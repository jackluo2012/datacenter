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

type ThemeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 主题列表
func NewThemeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThemeListLogic {
	return &ThemeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 主题列表: 页码默认值处理后转发(domain_id 为主题域过滤, 0 表示不过滤), 业务错误直接透传。
// 空列表仍 make 出空切片, 保证 JSON 序列化为 [] 而非 null。
func (l *ThemeListLogic) ThemeList(req *types.ThemeListReq) (resp *types.ThemeListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.ThemeRpc.ListThemes(l.ctx, &themeClient.ThemeListReq{
		Beid:     req.Beid,
		DomainId: req.DomainId,
		Page:     page,
		Size:     size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.ThemeInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *themeToTypes(item))
	}
	return &types.ThemeListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
