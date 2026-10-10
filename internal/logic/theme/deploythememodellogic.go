// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeployThemeModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeployThemeModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeployThemeModelLogic {
	return &DeployThemeModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeployThemeModelLogic) DeployThemeModel(req *types.ThemeIdReq) (resp *types.DeployThemeModelResp, err error) {
	// todo: add your logic here and delete this line

	return
}
