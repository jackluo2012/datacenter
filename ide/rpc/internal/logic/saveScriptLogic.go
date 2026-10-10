package logic

import (
	"context"

	"datacenter/ide/rpc/ide"
	"datacenter/ide/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveScriptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveScriptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveScriptLogic {
	return &SaveScriptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存脚本（新建或更新，更新自动 version+1）
func (l *SaveScriptLogic) SaveScript(in *ide.SaveScriptReq) (*ide.ScriptReply, error) {
	// todo: add your logic here and delete this line

	return &ide.ScriptReply{}, nil
}
