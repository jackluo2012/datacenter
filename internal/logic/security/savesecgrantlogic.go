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

type SaveSecGrantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存授权矩阵（主体→资源→许可级别）
func NewSaveSecGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSecGrantLogic {
	return &SaveSecGrantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存授权: 网关请求逐字段映射为 rpc 请求(同名字段一一对应), 转发后把授权记录转为网关类型返回。
// 契约差异: proto SaveGrantReq 无 Id 字段, 即 rpc 侧始终按新建处理(调整授权走删除后重建)。
func (l *SaveSecGrantLogic) SaveSecGrant(req *types.SaveSecGrantReq) (resp *types.SecGrant, err error) {
	reply, err := l.svcCtx.SecurityRpc.SaveGrant(l.ctx, &securityClient.SaveGrantReq{
		Subject:    req.Subject,
		Resource:   req.Resource,
		ResourceId: req.ResourceId,
		LevelCode:  req.LevelCode,
		ExpireAt:   req.ExpireAt,
		Beid:       req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return secGrantToTypes(reply.GetGrant()), nil
}
