// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/theme/rpc/themeClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshThemeMvLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 刷新物化视图（drop+create as select 真实执行）
func NewRefreshThemeMvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshThemeMvLogic {
	return &RefreshThemeMvLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 刷新物化视图: 网关请求逐字段映射为 rpc 请求后转发, rpc 侧 drop+create as select 真实执行并回写刷新状态。
// 契约差异: 网关 ThemeRefreshResp 与 rpc RefreshReply(ok/msg/cost_ms) 结构对齐, 逐字段回填;
// 刷新失败细节封装在 msg 中随正常响应返回, 不作为 error 抛出。
func (l *RefreshThemeMvLogic) RefreshThemeMv(req *types.ThemeIdReq) (resp *types.ThemeRefreshResp, err error) {
	reply, err := l.svcCtx.ThemeRpc.RefreshMaterializedView(l.ctx, &themeClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.ThemeRefreshResp{
		Ok:     reply.GetOk(),
		Msg:    reply.GetMsg(),
		CostMs: reply.GetCostMs(),
	}, nil
}
