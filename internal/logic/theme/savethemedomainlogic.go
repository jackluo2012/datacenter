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

type SaveThemeDomainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存主题域
func NewSaveThemeDomainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveThemeDomainLogic {
	return &SaveThemeDomainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存主题域: 网关请求逐字段映射为 rpc 请求, 转发后把主题域转为网关类型返回。
// 契约差异: 网关 SaveThemeDomainReq 无 Id 字段, rpc SaveDomainReq.Id 恒为 0, 即 rpc 侧始终按新建处理。
func (l *SaveThemeDomainLogic) SaveThemeDomain(req *types.SaveThemeDomainReq) (resp *types.ThemeDomainInfo, err error) {
	reply, err := l.svcCtx.ThemeRpc.SaveDomain(l.ctx, &themeClient.SaveDomainReq{
		Code:    req.Code,
		Name:    req.Name,
		Comment: req.Comment,
		Beid:    req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return themeDomainToTypes(reply.GetDomain()), nil
}
