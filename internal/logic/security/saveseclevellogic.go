// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveSecLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveSecLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSecLevelLogic {
	return &SaveSecLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveSecLevelLogic) SaveSecLevel(req *types.SaveSecLevelReq) (resp *types.SecLevelInfo, err error) {
	// todo: add your logic here and delete this line

	return
}
