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

type DeleteSecGrantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除授权
func NewDeleteSecGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSecGrantLogic {
	return &DeleteSecGrantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除授权: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteGrant 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteSecGrantLogic) DeleteSecGrant(req *types.SecIdReq) error {
	_, err := l.svcCtx.SecurityRpc.DeleteGrant(l.ctx, &securityClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
