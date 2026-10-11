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

type DeleteMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMenuLogic {
	return &DeleteMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除菜单：有子菜单拒删、被角色引用拒删；通过校验后同事务删除菜单行并级联清理角色引用
func (l *DeleteMenuLogic) DeleteMenu(in *sys.IdReq) (*sys.IdReq, error) {
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	// 存在性校验（含跨租户不可见）
	if _, err := l.svcCtx.MenuModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrMenuNotFound
		}
		return nil, err
	}
	// 有子菜单拒删（避免子节点成孤儿）
	children, err := l.svcCtx.MenuModel.CountChildren(l.ctx, in.Beid, in.Id)
	if err != nil {
		return nil, err
	}
	if children > 0 {
		return nil, ErrMenuHasChildren
	}
	// 被角色引用拒删（防角色权限静默收窄）
	roleCount, err := l.svcCtx.RoleModel.CountByMenu(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if roleCount > 0 {
		return nil, ErrMenuInUse
	}
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 级联清理角色引用（并发绑定窗口内可能新出现引用行，删除为幂等兜底）
		if err := l.svcCtx.RoleModel.RemoveMenuRef(ctx, session, in.Id); err != nil {
			return err
		}
		return l.svcCtx.MenuModel.Delete(ctx, session, in.Id, in.Beid)
	})
	if err != nil {
		return nil, err
	}
	l.Infof("删除菜单成功 id=%d beid=%d", in.Id, in.Beid)
	return in, nil
}
