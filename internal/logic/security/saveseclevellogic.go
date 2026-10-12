// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/security/rpc/securityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveSecLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存分级目录（L1-L4 及自定义级别）
func NewSaveSecLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSecLevelLogic {
	return &SaveSecLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存分级: 网关请求逐字段映射为 rpc 请求, 转发后把分级转为网关类型返回。
// 契约差异: 网关 SaveSecLevelReq 无 Id 字段, rpc SaveLevelReq.Id 恒为 0, 即 rpc 侧始终按新建处理。
func (l *SaveSecLevelLogic) SaveSecLevel(req *types.SaveSecLevelReq) (resp *types.SecLevelInfo, err error) {
	reply, err := l.svcCtx.SecurityRpc.SaveLevel(l.ctx, &securityClient.SaveLevelReq{
		Code:    req.Code,
		Name:    req.Name,
		Comment: req.Comment,
		Beid:    req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return secLevelToTypes(reply.GetLevel()), nil
}
