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

type SaveScriptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存脚本
func NewSaveScriptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveScriptLogic {
	return &SaveScriptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存脚本: 网关请求逐字段映射为 rpc 请求(id=0 新建, >0 更新自动 version+1), 转发后把 Script 转为网关类型返回
func (l *SaveScriptLogic) SaveScript(req *types.SaveScriptReq) (resp *types.Script, err error) {
	reply, err := l.svcCtx.IdeRpc.SaveScript(l.ctx, &ideClient.SaveScriptReq{
		Id:           req.Id,
		Name:         req.Name,
		SqlText:      req.SqlText,
		DatasourceId: req.DatasourceId,
		DbName:       req.DbName,
		Beid:         req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return scriptToTypes(reply.Script), nil
}
