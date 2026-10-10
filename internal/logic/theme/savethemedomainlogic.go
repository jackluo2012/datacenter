// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveThemeDomainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveThemeDomainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveThemeDomainLogic {
	return &SaveThemeDomainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveThemeDomainLogic) SaveThemeDomain(req *types.SaveThemeDomainReq) (resp *types.ThemeDomainInfo, err error) {
	// todo: add your logic here and delete this line

	return
}
