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

type SysRoleInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleInfoLogic {
	return &SysRoleInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 角色详情: 返回含 menu_ids(角色菜单授权), beid 缺省取 JWT ctx, 转发 sys.rpc GetRole
func (l *SysRoleInfoLogic) SysRoleInfo(req *types.SysRoleIdReq) (*types.SysRoleResp, error) {
	reply, err := l.svcCtx.SysRpc.GetRole(l.ctx, &sysclient.IdReq{
		Id:   req.Id,
		Beid: beidOrCtx(req.Beid, l.ctx),
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.SysRoleResp{Role: *sysRoleToTypes(reply.GetRole())}, nil
}
