package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveLevelLogic {
	return &SaveLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分级目录
func (l *SaveLevelLogic) SaveLevel(in *security.SaveLevelReq) (*security.LevelReply, error) {
	// todo: add your logic here and delete this line

	return &security.LevelReply{}, nil
}
