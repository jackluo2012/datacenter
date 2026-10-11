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

type SysRoleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleUpdateLogic {
	return &SysRoleUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 更新角色: Id 必传走更新语义, beid 缺省取 JWT ctx, 转发 sys.rpc UpdateRole
func (l *SysRoleUpdateLogic) SysRoleUpdate(req *types.SysRoleSaveReq) (*types.SysRoleResp, error) {
	reply, err := l.svcCtx.SysRpc.UpdateRole(l.ctx, &sysclient.SaveSysRoleReq{
		Id:       req.Id,
		RoleKey:  req.RoleKey,
		RoleName: req.RoleName,
		Status:   req.Status,
		Remark:   req.Remark,
		MenuIds:  req.MenuIds,
		Beid:     beidOrCtx(req.Beid, l.ctx),
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.SysRoleResp{Role: *sysRoleToTypes(reply.GetRole())}, nil
}
