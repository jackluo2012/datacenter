package logic

import (
	"context"

	"datacenter/ide/rpc/ide"
	"datacenter/ide/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunScriptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunScriptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunScriptLogic {
	return &RunScriptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行脚本（开发调试：仅 SELECT，行数保护截断）
func (l *RunScriptLogic) RunScript(in *ide.RunScriptReq) (*ide.RunScriptReply, error) {
	// todo: add your logic here and delete this line

	return &ide.RunScriptReply{}, nil
}
