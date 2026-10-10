// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ThemeDomainListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewThemeDomainListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThemeDomainListLogic {
	return &ThemeDomainListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ThemeDomainListLogic) ThemeDomainList(req *types.ThemeIdReq) (resp *types.ThemeDomainListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
