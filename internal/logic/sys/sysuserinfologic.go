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

type SysUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserInfoLogic {
	return &SysUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 管理用户详情: beid 缺省取 JWT ctx, 转发 sys.rpc GetUser
func (l *SysUserInfoLogic) SysUserInfo(req *types.SysUserIdReq) (*types.SysUserResp, error) {
	reply, err := l.svcCtx.SysRpc.GetUser(l.ctx, &sysclient.IdReq{
		Id:   req.Id,
		Beid: beidOrCtx(req.Beid, l.ctx),
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.SysUserResp{User: *sysUserToTypes(reply.GetUser())}, nil
}
