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

type DeleteThemeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除主题
func NewDeleteThemeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteThemeLogic {
	return &DeleteThemeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除主题: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteTheme 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteThemeLogic) DeleteTheme(req *types.ThemeIdReq) error {
	_, err := l.svcCtx.ThemeRpc.DeleteTheme(l.ctx, &themeClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
