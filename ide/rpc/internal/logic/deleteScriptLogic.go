package logic

import (
	"context"
	"errors"

	"datacenter/ide/model"
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

// 删除脚本：先确认存在且归属当前租户（避免误报删除成功），再按 id+beid 定位删除
func (l *DeleteScriptLogic) DeleteScript(in *ide.IdReq) (*ide.IdReq, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrScriptNotFound
	}
	if _, err := l.svcCtx.ScriptModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrScriptNotFound
		}
		return nil, err
	}
	if err := l.svcCtx.ScriptModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	l.Infof("[DeleteScript] id=%d beid=%d", in.Id, in.Beid)
	return &ide.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
