// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ThemeMvListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewThemeMvListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThemeMvListLogic {
	return &ThemeMvListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ThemeMvListLogic) ThemeMvList(req *types.ThemeMvListReq) (resp *types.ThemeMvListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
