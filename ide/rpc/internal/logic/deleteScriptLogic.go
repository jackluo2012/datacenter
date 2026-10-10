package logic

import (
	"context"

	"datacenter/ide/rpc/ide"
	"datacenter/ide/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteScriptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteScriptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteScriptLogic {
	return &DeleteScriptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除脚本
func (l *DeleteScriptLogic) DeleteScript(in *ide.IdReq) (*ide.IdReq, error) {
	// todo: add your logic here and delete this line

	return &ide.IdReq{}, nil
}
