// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package datasvc

import (
	"context"

	"datacenter/datasvc/rpc/dataSvcClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterAppLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 注册调用应用
func NewRegisterAppLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterAppLogic {
	return &RegisterAppLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// RegisterApp 转发 datasvc rpc 注册调用应用，
// 把返回的 app_id / app_key / app_secret 映射为网关 AppResp（secret 仅此一次）
func (l *RegisterAppLogic) RegisterApp(req *types.RegisterAppReq) (resp *types.AppResp, err error) {
	// 逐字段映射为 rpc 请求
	rpcResp, err := l.svcCtx.DataSvcRpc.RegisterApp(l.ctx, &dataSvcClient.RegisterAppReq{
		Name:         req.Name,
		RateLimitQps: req.RateLimitQps,
		IpWhitelist:  req.IpWhitelist,
		Beid:         req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	// rpc 返回 ApiApp（含 id/app_key）+ AppSecret，映射为网关 AppResp
	return &types.AppResp{
		AppId:     rpcResp.GetApp().GetId(),
		AppKey:    rpcResp.GetApp().GetAppKey(),
		AppSecret: rpcResp.GetAppSecret(),
	}, nil
}
