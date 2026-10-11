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

type SysGetroutersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysGetroutersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysGetroutersLogic {
	return &SysGetroutersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 前端动态路由: 网关从 JWT ctx 取 uid/beid 传入, 替代样板前端 auth store 会话,
// 转发 sys.rpc GetRouters
func (l *SysGetroutersLogic) SysGetrouters() (*types.SysRoutersResp, error) {
	reply, err := l.svcCtx.SysRpc.GetRouters(l.ctx, &sysclient.GetRoutersReq{
		Uid:  ctxUid(l.ctx),
		Beid: ctxBeid(l.ctx),
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	routers := make([]types.SysRouterBranch, 0, len(reply.GetRouters()))
	for _, item := range reply.GetRouters() {
		routers = append(routers, *routerBranchToTypes(item))
	}
	return &types.SysRoutersResp{Routers: routers}, nil
}
