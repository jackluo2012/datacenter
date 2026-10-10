package logic

import (
	"context"

	"datacenter/ide/rpc/ide"
	"datacenter/ide/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListScriptsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListScriptsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListScriptsLogic {
	return &ListScriptsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 脚本列表
func (l *ListScriptsLogic) ListScripts(in *ide.ScriptListReq) (*ide.ScriptListReply, error) {
	// todo: add your logic here and delete this line

	return &ide.ScriptListReply{}, nil
}
