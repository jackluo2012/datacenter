package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLevelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLevelsLogic {
	return &ListLevelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListLevelsLogic) ListLevels(in *security.IdReq) (*security.LevelListReply, error) {
	// todo: add your logic here and delete this line

	return &security.LevelListReply{}, nil
}
