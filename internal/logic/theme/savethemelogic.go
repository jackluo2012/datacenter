// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveThemeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveThemeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveThemeLogic {
	return &SaveThemeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveThemeLogic) SaveTheme(req *types.SaveThemeReq) (resp *types.ThemeInfo, err error) {
	// todo: add your logic here and delete this line

	return
}
