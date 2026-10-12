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

type SaveThemeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存主题
func NewSaveThemeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveThemeLogic {
	return &SaveThemeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存主题: 网关请求逐字段映射为 rpc 请求, 转发后把主题转为网关类型返回。
// 契约差异: 网关 SaveThemeReq 无 Id 字段, rpc SaveThemeReq.Id 恒为 0, 即 rpc 侧始终按新建处理。
func (l *SaveThemeLogic) SaveTheme(req *types.SaveThemeReq) (resp *types.ThemeInfo, err error) {
	reply, err := l.svcCtx.ThemeRpc.SaveTheme(l.ctx, &themeClient.SaveThemeReq{
		DomainId: req.DomainId,
		Code:     req.Code,
		Name:     req.Name,
		Comment:  req.Comment,
		Beid:     req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return themeToTypes(reply.GetTheme()), nil
}
