package logic

import (
	"context"

	"datacenter/ide/rpc/ide"
	"datacenter/ide/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScriptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetScriptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScriptLogic {
	return &GetScriptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 脚本详情
func (l *GetScriptLogic) GetScript(in *ide.IdReq) (*ide.ScriptReply, error) {
	// todo: add your logic here and delete this line

	return &ide.ScriptReply{}, nil
}
