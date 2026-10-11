package logic

import (
	"context"
	"errors"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 角色详情：角色信息 + menu_ids（菜单授权回显）
func (l *GetRoleLogic) GetRole(in *sys.IdReq) (*sys.SysRoleReply, error) {
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	role, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	menuIds, err := l.svcCtx.RoleModel.GetMenuIds(l.ctx, role.Id)
	if err != nil {
		return nil, err
	}
	return &sys.SysRoleReply{Role: roleToPb(role, menuIds)}, nil
}
