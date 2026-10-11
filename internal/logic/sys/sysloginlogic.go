// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package sys

import (
	"context"
	"time"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	sysclient "datacenter/sys/rpc/sysClient"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type SysLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理台登录：sys-rpc 校验账密（bcrypt），通过后网关签发 JWT（uid= sys_user.id）。
func NewSysLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysLoginLogic {
	return &SysLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysLoginLogic) SysLogin(req *types.SysLoginReq) (resp *types.SysLoginResp, err error) {
	reply, err := l.svcCtx.SysRpc.SysLogin(l.ctx, &sysclient.SysLoginReq{
		Username: req.Username,
		Password: req.Password,
		Beid:     req.Beid,
	})
	if err != nil {
		return nil, err //业务错误直接透传 rpc error
	}

	token, err := l.getJwtToken(reply.Uid, req.Beid)
	if err != nil {
		return nil, err
	}

	roles := make([]string, 0, len(reply.Roles))
	for _, r := range reply.Roles {
		roles = append(roles, r)
	}
	perms := make([]string, 0, len(reply.Perms))
	for _, p := range reply.Perms {
		perms = append(perms, p)
	}

	return &types.SysLoginResp{
		Token:    token,
		Uid:      reply.Uid,
		Username: reply.Username,
		Nickname: reply.Nickname,
		Roles:    roles,
		Perms:    perms,
	}, nil
}

// getJwtToken 管理台 JWT 签发（uid 对应 sys_user.id，与 C 端 base_member 的 uid 空间隔离）
func (l *SysLoginLogic) getJwtToken(uid, beid int64) (string, error) {
	iat := time.Now().Unix()
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + l.svcCtx.Config.Auth.AccessExpire
	claims["iat"] = iat
	claims["uid"] = uid
	claims["beid"] = beid
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
}
