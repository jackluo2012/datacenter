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

type ScriptInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 脚本详情
func NewScriptInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScriptInfoLogic {
	return &ScriptInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 脚本详情: 网关请求逐字段映射为 rpc 请求, 转发后把 Script 转为网关类型返回
func (l *ScriptInfoLogic) ScriptInfo(req *types.ScriptIdReq) (resp *types.Script, err error) {
	reply, err := l.svcCtx.IdeRpc.GetScript(l.ctx, &ideClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return scriptToTypes(reply.Script), nil
}
