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

type SysUserUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserUpdateLogic {
	return &SysUserUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 更新管理用户: password 为空表示不改密码, beid 缺省取 JWT ctx, 转发 sys.rpc UpdateUser
func (l *SysUserUpdateLogic) SysUserUpdate(req *types.SysUserUpdateReq) (*types.SysUserResp, error) {
	reply, err := l.svcCtx.SysRpc.UpdateUser(l.ctx, &sysclient.UpdateSysUserReq{
		Id:       req.Id,
		Nickname: req.Nickname,
		Mobile:   req.Mobile,
		Email:    req.Email,
		Status:   req.Status,
		Remark:   req.Remark,
		Password: req.Password,
		RoleIds:  req.RoleIds,
		Beid:     beidOrCtx(req.Beid, l.ctx),
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.SysUserResp{User: *sysUserToTypes(reply.GetUser())}, nil
}
