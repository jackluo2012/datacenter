// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveThemeMvLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveThemeMvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveThemeMvLogic {
	return &SaveThemeMvLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveThemeMvLogic) SaveThemeMv(req *types.SaveThemeMvReq) (resp *types.ThemeMvInfo, err error) {
	// todo: add your logic here and delete this line

	return
}
