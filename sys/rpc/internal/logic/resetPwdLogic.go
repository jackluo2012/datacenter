package logic

import (
	"context"
	"errors"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPwdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetPwdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPwdLogic {
	return &ResetPwdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 重置密码：新密码 bcrypt 哈希后落库（beid>0 时限定租户）
func (l *ResetPwdLogic) ResetPwd(in *sys.ResetPwdReq) (*sys.IdReq, error) {
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	if in.Password == "" {
		return nil, ErrPasswordRequired
	}
	// 存在性校验（含跨租户不可见）
	if _, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	hash, err := model.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	if err = l.svcCtx.UserModel.UpdatePwd(l.ctx, nil, in.Id, hash, in.Beid); err != nil {
		return nil, err
	}
	l.Infof("重置密码成功 id=%d beid=%d", in.Id, in.Beid)
	return &sys.IdReq{Id: in.Id, Beid: in.Beid}, nil
}
