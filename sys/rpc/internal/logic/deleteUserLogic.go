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

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除用户：同事务清理用户行与角色绑定
func (l *DeleteUserLogic) DeleteUser(in *sys.IdReq) (*sys.IdReq, error) {
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	// 存在性校验（含跨租户不可见）
	if _, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	err := l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先清角色绑定再删用户（SetUserRoles(id, nil) 语义为清空绑定）
		if err := l.svcCtx.UserModel.SetUserRoles(ctx, session, in.Id, nil); err != nil {
			return err
		}
		return l.svcCtx.UserModel.Delete(ctx, session, in.Id, in.Beid)
	})
	if err != nil {
		return nil, err
	}
	l.Infof("删除用户成功 id=%d beid=%d", in.Id, in.Beid)
	return in, nil
}
