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

// 脚本列表：beid 多租户必带，keyword 模糊匹配 name，分页返回列表与总数
func (l *ListScriptsLogic) ListScripts(in *ide.ScriptListReq) (*ide.ScriptListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, total, err := l.svcCtx.ScriptModel.FindList(l.ctx, in.Beid, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	reply := &ide.ScriptListReply{Total: total}
	for _, s := range list {
		reply.List = append(reply.List, scriptToPb(s))
	}
	return reply, nil
}
