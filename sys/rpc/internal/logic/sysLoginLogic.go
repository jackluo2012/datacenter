package logic

import (
	"context"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理台登录：sys_user（bcrypt 校验），与 C 端 base_member 登录（/user/login）分离。
// 失败统一返回"用户名或密码错误"，不泄露账号存在性。
func NewSysLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysLoginLogic {
	return &SysLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysLoginLogic) SysLogin(in *sys.SysLoginReq) (*sys.SysLoginReply, error) {
	if in.Username == "" || in.Password == "" {
		return nil, ErrUserOrPwdWrong
	}
	user, err := l.svcCtx.UserModel.FindByUsername(l.ctx, in.Username)
	if err != nil {
		if model.ErrNotFound == err {
			// 空库兜底：首次登录前播种 RBAC（幂等），种子后重查一次（解决"登录前无入口触发播种"的先有鸡问题）
			if serr := seedRBAC(l.ctx, l.svcCtx, in.Beid); serr != nil {
				l.Logger.Errorf("[定位] RBAC 播种失败: %v", serr)
			} else {
				l.Logger.Infof("[定位] RBAC 播种完成")
			}
			user, err = l.svcCtx.UserModel.FindByUsername(l.ctx, in.Username)
			if err != nil {
				return nil, ErrUserOrPwdWrong
			}
		} else {
			return nil, err
		}
	}
	if user.Status != 1 {
		return nil, ErrUserDisabled
	}
	if !model.CheckPassword(user.Password, in.Password) {
		return nil, ErrUserOrPwdWrong
	}
	// 角色 key 集 + 按钮权限标识集（登录即携带，前端 v-perm 与动态菜单依赖）
	roles, err := l.svcCtx.RoleModel.FindKeysByUser(l.ctx, user.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	perms, err := l.svcCtx.MenuModel.FindPermsByUid(l.ctx, user.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &sys.SysLoginReply{
		Uid:      user.Id,
		Username: user.Username,
		Nickname: user.Nickname,
		Roles:    roles,
		Perms:    perms,
	}, nil
}
