package logic

import (
	"context"
	"errors"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInfoLogic {
	return &GetInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登录增强：用户信息 + roles（角色 key 集）+ perms（按钮权限标识集）+ 菜单树（目录+菜单，按钮剔除）。
// uid 由网关从 JWT 解出后传入，rpc 内不做 JWT 校验（信任网关）
func (l *GetInfoLogic) GetInfo(in *sys.GetInfoReq) (*sys.GetInfoReply, error) {
	if in.Uid <= 0 {
		return nil, ErrIdRequired
	}
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Uid, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 角色 key 集（前端按钮权限指令 / 隐藏入口用）
	roles, err := l.svcCtx.RoleModel.FindKeysByUser(l.ctx, in.Uid, in.Beid)
	if err != nil {
		return nil, err
	}

	// 用户可见菜单（含按钮）：先建树收集 perm 标识集，再剔除按钮得展示菜单树
	menus, err := l.svcCtx.MenuModel.FindByUid(l.ctx, in.Uid, in.Beid)
	if err != nil {
		return nil, err
	}
	if len(menus) == 0 {
		// 兜底播种：全局菜单为空（首次部署直调 GetInfo）时初始化种子，可重入
		all, err := l.svcCtx.MenuModel.FindAll(l.ctx, in.Beid)
		if err != nil {
			return nil, err
		}
		if len(all) == 0 {
			if err = seedRBAC(l.ctx, l.svcCtx, in.Beid); err != nil {
				return nil, err
			}
			if menus, err = l.svcCtx.MenuModel.FindByUid(l.ctx, in.Uid, in.Beid); err != nil {
				return nil, err
			}
		}
	}
	tree := BuildMenuTree(menusToPb(menus))
	perms := CollectPerms(tree)

	return &sys.GetInfoReply{
		Uid:      user.Id,
		Username: user.Username,
		Nickname: user.Nickname,
		Roles:    roles,
		Perms:    perms,
		Menus:    StripButtons(tree),
	}, nil
}
