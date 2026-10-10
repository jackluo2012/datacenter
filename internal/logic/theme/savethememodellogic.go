// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveThemeModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveThemeModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveThemeModelLogic {
	return &SaveThemeModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveThemeModelLogic) SaveThemeModel(req *types.SaveThemeModelReq) (resp *types.ThemeModelInfo, err error) {
	// todo: add your logic here and delete this line

	return
}
