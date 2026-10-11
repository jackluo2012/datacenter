package logic

import (
	"context"
	"errors"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改角色：基础字段整存 + role_key 查重（排除自身）+ 菜单绑定全量重置（先删后插）
func (l *UpdateRoleLogic) UpdateRole(in *sys.SaveSysRoleReq) (*sys.SysRoleReply, error) {
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	// 存在性校验（含跨租户不可见）
	old, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	// role_key 查重：同 beid 下已被其他角色占用则拒绝
	if dup, err := l.svcCtx.RoleModel.FindByKey(l.ctx, in.Beid, in.RoleKey); err == nil && dup.Id != old.Id {
		return nil, ErrRoleKeyDuplicate
	} else if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	menuIds := dedupeInt64s(in.MenuIds)
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err = l.svcCtx.RoleModel.Update(ctx, session, &model.Role{
			Id:       old.Id,
			RoleKey:  in.RoleKey,
			RoleName: in.RoleName,
			Status:   int64(in.Status),
			Remark:   in.Remark,
			Beid:     old.Beid,
		}); err != nil {
			return err
		}
		return l.svcCtx.RoleModel.SetRoleMenus(ctx, session, old.Id, menuIds)
	})
	if err != nil {
		return nil, err
	}

	role, err := l.svcCtx.RoleModel.FindOne(l.ctx, old.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	l.Infof("修改角色成功 id=%d menus=%d", old.Id, len(menuIds))
	return &sys.SysRoleReply{Role: roleToPb(role, menuIds)}, nil
}
