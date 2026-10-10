// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteThemeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteThemeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteThemeLogic {
	return &DeleteThemeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteThemeLogic) DeleteTheme(req *types.ThemeIdReq) error {
	// todo: add your logic here and delete this line

	return nil
}
