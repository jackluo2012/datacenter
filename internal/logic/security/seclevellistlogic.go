// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SecLevelListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSecLevelListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecLevelListLogic {
	return &SecLevelListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SecLevelListLogic) SecLevelList(req *types.SecIdReq) (resp *types.SecLevelListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
