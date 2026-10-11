// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package sys

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	sysclient "datacenter/sys/rpc/sysClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleDeleteLogic {
	return &SysRoleDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除角色: beid 缺省取 JWT ctx, 转发 sys.rpc DeleteRole
func (l *SysRoleDeleteLogic) SysRoleDelete(req *types.SysRoleIdReq) error {
	_, err := l.svcCtx.SysRpc.DeleteRole(l.ctx, &sysclient.IdReq{
		Id:   req.Id,
		Beid: beidOrCtx(req.Beid, l.ctx),
	})
	// 业务错误直接透传 rpc error
	return err
}
