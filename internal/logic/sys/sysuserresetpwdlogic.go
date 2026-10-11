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

type SysUserResetpwdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysUserResetpwdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserResetpwdLogic {
	return &SysUserResetpwdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 重置管理用户密码: beid 缺省取 JWT ctx, 转发 sys.rpc ResetPwd
func (l *SysUserResetpwdLogic) SysUserResetpwd(req *types.SysUserResetPwdReq) error {
	_, err := l.svcCtx.SysRpc.ResetPwd(l.ctx, &sysclient.ResetPwdReq{
		Id:       req.Id,
		Password: req.Password,
		Beid:     beidOrCtx(req.Beid, l.ctx),
	})
	// 业务错误直接透传 rpc error
	return err
}
