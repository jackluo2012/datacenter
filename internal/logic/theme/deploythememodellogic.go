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

type DeployThemeModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 模型部署（按模型在目标库生成建表 DDL 并执行）
func NewDeployThemeModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeployThemeModelLogic {
	return &DeployThemeModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 模型部署: 网关请求逐字段映射为 rpc 请求后转发, rpc 侧生成建表 DDL 并在目标库真实执行。
// 契约差异: 网关 DeployThemeModelResp 与 rpc RefreshReply(ok/msg/cost_ms) 结构对齐, 逐字段回填;
// 部署失败细节(如 DDL 报错)封装在 msg 中随正常响应返回, 不作为 error 抛出。
func (l *DeployThemeModelLogic) DeployThemeModel(req *types.ThemeIdReq) (resp *types.DeployThemeModelResp, err error) {
	reply, err := l.svcCtx.ThemeRpc.DeployThemeModel(l.ctx, &themeClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.DeployThemeModelResp{
		Ok:     reply.GetOk(),
		Msg:    reply.GetMsg(),
		CostMs: reply.GetCostMs(),
	}, nil
}
