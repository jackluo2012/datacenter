package logic

import (
	"context"
	"errors"

	"datacenter/ide/model"
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

// 脚本详情（FindOne 已带多租户过滤，越权/不存在统一按脚本不存在处理）
func (l *GetScriptLogic) GetScript(in *ide.IdReq) (*ide.ScriptReply, error) {
	if in.Id <= 0 {
		return nil, ErrScriptNotFound
	}
	s, err := l.svcCtx.ScriptModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrScriptNotFound
		}
		return nil, err
	}
	return &ide.ScriptReply{Script: scriptToPb(s)}, nil
}
