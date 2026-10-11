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

type DeleteRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除角色：有用户引用拒绝；同事务清理角色行与菜单绑定
func (l *DeleteRoleLogic) DeleteRole(in *sys.IdReq) (*sys.IdReq, error) {
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	// 存在性校验（含跨租户不可见）
	if _, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	// 有用户引用拒删（内置 admin/viewer 有账号引用时同样被拦，防误删导致权限悬空）
	userCount, err := l.svcCtx.UserModel.CountByRole(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if userCount > 0 {
		return nil, ErrRoleInUse
	}
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先清菜单绑定再删角色（SetRoleMenus(id, nil) 语义为清空绑定）
		if err := l.svcCtx.RoleModel.SetRoleMenus(ctx, session, in.Id, nil); err != nil {
			return err
		}
		return l.svcCtx.RoleModel.Delete(ctx, session, in.Id, in.Beid)
	})
	if err != nil {
		return nil, err
	}
	l.Infof("删除角色成功 id=%d beid=%d", in.Id, in.Beid)
	return in, nil
}
