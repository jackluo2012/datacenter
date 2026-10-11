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

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改用户：基础字段整存（username 不可变）；password 非空时同步重置；
// role_ids 全量重置用户角色绑定（先删后插）
func (l *UpdateUserLogic) UpdateUser(in *sys.UpdateSysUserReq) (*sys.SysUserReply, error) {
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	// 存在性校验（含跨租户不可见）
	old, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 密码非空则重置（bcrypt），空 = 不改
	var passwordHash string
	if in.Password != "" {
		if passwordHash, err = model.HashPassword(in.Password); err != nil {
			return nil, err
		}
	}

	roleIds := dedupeInt64s(in.RoleIds)
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err = l.svcCtx.UserModel.Update(ctx, session, &model.User{
			Id:       old.Id,
			Nickname: in.Nickname,
			Mobile:   in.Mobile,
			Email:    in.Email,
			Status:   int64(in.Status),
			Remark:   in.Remark,
			Beid:     old.Beid,
		}); err != nil {
			return err
		}
		// 密码非空时同事务重置
		if passwordHash != "" {
			if err = l.svcCtx.UserModel.UpdatePwd(ctx, session, old.Id, passwordHash, 0); err != nil {
				return err
			}
		}
		// 角色绑定全量重置（先删后插）
		return l.svcCtx.UserModel.SetUserRoles(ctx, session, old.Id, roleIds)
	})
	if err != nil {
		return nil, err
	}

	user, err := l.svcCtx.UserModel.FindOne(l.ctx, old.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	l.Infof("修改用户成功 id=%d 重置密码=%v roles=%v", old.Id, passwordHash != "", roleIds)
	return &sys.SysUserReply{User: userToPb(user, roleIds)}, nil
}
