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

type SysRoleCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleCreateLogic {
	return &SysRoleCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 创建角色: 强制 Id=0 走新建语义, beid 缺省取 JWT ctx, 转发 sys.rpc CreateRole
func (l *SysRoleCreateLogic) SysRoleCreate(req *types.SysRoleSaveReq) (*types.SysRoleResp, error) {
	reply, err := l.svcCtx.SysRpc.CreateRole(l.ctx, &sysclient.SaveSysRoleReq{
		Id:       0,
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
