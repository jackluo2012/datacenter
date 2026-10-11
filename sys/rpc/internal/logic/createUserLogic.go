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

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 新增用户：登录名查重 → bcrypt 哈希 → 事务内落用户 + 角色绑定
func (l *CreateUserLogic) CreateUser(in *sys.CreateSysUserReq) (*sys.SysUserReply, error) {
	username := strings.TrimSpace(in.Username)
	if username == "" {
		return nil, ErrUsernameRequired
	}
	if in.Password == "" {
		return nil, ErrPasswordRequired
	}
	// 登录名查重（uk_username 全局唯一）
	_, err := l.svcCtx.UserModel.FindByUsername(l.ctx, username)
	switch {
	case err == nil:
		return nil, ErrUsernameDuplicate
	case !errors.Is(err, model.ErrNotFound):
		return nil, err
	}

	// bcrypt 哈希（golang.org/x/crypto 依赖树内，见 model/password.go 说明）
	hash, err := model.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	roleIds := dedupeInt64s(in.RoleIds)
	var userId int64
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		userId, err = l.svcCtx.UserModel.Insert(ctx, session, &model.User{
			Username: username,
			Password: hash,
			Nickname: in.Nickname,
			Mobile:   in.Mobile,
			Email:    in.Email,
			Status:   model.StatusEnabled, // 新建默认启用
			Remark:   in.Remark,
			Beid:     in.Beid,
		})
		if err != nil {
			return err
		}
		// 角色绑定：先删后插（新用户无历史绑定，整存语义）
		return l.svcCtx.UserModel.SetUserRoles(ctx, session, userId, roleIds)
	})
	if err != nil {
		return nil, err
	}

	user, err := l.svcCtx.UserModel.FindOne(l.ctx, userId, 0)
	if err != nil {
		return nil, err
	}
	l.Infof("新增用户成功 id=%d username=%s roles=%v", userId, username, roleIds)
	return &sys.SysUserReply{User: userToPb(user, roleIds)}, nil
}
