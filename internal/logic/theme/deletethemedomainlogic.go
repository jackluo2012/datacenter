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

type DeleteThemeDomainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除主题域
func NewDeleteThemeDomainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteThemeDomainLogic {
	return &DeleteThemeDomainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除主题域: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteDomain 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteThemeDomainLogic) DeleteThemeDomain(req *types.ThemeIdReq) error {
	_, err := l.svcCtx.ThemeRpc.DeleteDomain(l.ctx, &themeClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
