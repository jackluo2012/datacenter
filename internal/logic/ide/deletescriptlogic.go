// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package ide

import (
	"context"

	"datacenter/ide/rpc/ideClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteScriptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除脚本
func NewDeleteScriptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteScriptLogic {
	return &DeleteScriptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除脚本: 网关请求逐字段映射为 rpc 请求后转发, 业务错误直接透传
func (l *DeleteScriptLogic) DeleteScript(req *types.ScriptIdReq) error {
	_, err := l.svcCtx.IdeRpc.DeleteScript(l.ctx, &ideClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
