// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package sys

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	sysclient "datacenter/sys/rpc/sysClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysUserCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserCreateLogic {
	return &SysUserCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 创建管理用户: beid 缺省取 JWT ctx, 转发 sys.rpc CreateUser, 业务错误直接透传
func (l *SysUserCreateLogic) SysUserCreate(req *types.SysUserCreateReq) (*types.SysUserResp, error) {
	reply, err := l.svcCtx.SysRpc.CreateUser(l.ctx, &sysclient.CreateSysUserReq{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Mobile:   req.Mobile,
		Email:    req.Email,
		Remark:   req.Remark,
		RoleIds:  req.RoleIds,
		Beid:     beidOrCtx(req.Beid, l.ctx),
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.SysUserResp{User: *sysUserToTypes(reply.GetUser())}, nil
}
