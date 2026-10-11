package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 新增角色：role_key 查重 → 事务内落角色 + 菜单绑定（先删后插）
func (l *CreateRoleLogic) CreateRole(in *sys.SaveSysRoleReq) (*sys.SysRoleReply, error) {
	roleKey := strings.TrimSpace(in.RoleKey)
	if roleKey == "" {
		return nil, ErrRoleKeyRequired
	}
	if strings.TrimSpace(in.RoleName) == "" {
		return nil, ErrNameRequired
	}
	// role_key 查重（uk_beid_key）
	_, err := l.svcCtx.RoleModel.FindByKey(l.ctx, in.Beid, roleKey)
	switch {
	case err == nil:
		return nil, ErrRoleKeyDuplicate
	case !errors.Is(err, model.ErrNotFound):
		return nil, err
	}

	menuIds := dedupeInt64s(in.MenuIds)
	var roleId int64
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		roleId, err = l.svcCtx.RoleModel.Insert(ctx, session, &model.Role{
			RoleKey:  roleKey,
			RoleName: in.RoleName,
			Status:   statusOrDefault(int64(in.Status)),
			Remark:   in.Remark,
			Beid:     in.Beid,
		})
		if err != nil {
			return err
		}
		return l.svcCtx.RoleModel.SetRoleMenus(ctx, session, roleId, menuIds)
	})
	if err != nil {
		return nil, err
	}

	role, err := l.svcCtx.RoleModel.FindOne(l.ctx, roleId, 0)
	if err != nil {
		return nil, err
	}
	l.Infof("新增角色成功 id=%d role_key=%s menus=%d", roleId, roleKey, len(menuIds))
	return &sys.SysRoleReply{Role: roleToPb(role, menuIds)}, nil
}

// statusOrDefault 状态缺省：未传（0）默认启用；显式传 0 停用场景由修改接口承接，
// 新建语义默认启用更安全（与 proto 缺省值 0 的歧义规避）
func statusOrDefault(status int64) int64 {
	if status == 0 {
		return model.StatusEnabled
	}
	return status
}
