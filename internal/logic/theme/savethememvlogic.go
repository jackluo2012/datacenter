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

type SaveThemeMvLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存物化视图
func NewSaveThemeMvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveThemeMvLogic {
	return &SaveThemeMvLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存物化视图: 网关请求逐字段映射为 rpc 请求(仅登记元数据, 不触发刷新), 转发后把物化视图转为网关类型返回。
// 契约差异: 网关 SaveThemeMvReq 对应 rpc SaveMvReq(无 Id 字段, 即 rpc 侧始终按新建处理);
// last_status/last_refresh_at 由刷新动作回写, 新建记录为空串/0。
func (l *SaveThemeMvLogic) SaveThemeMv(req *types.SaveThemeMvReq) (resp *types.ThemeMvInfo, err error) {
	reply, err := l.svcCtx.ThemeRpc.SaveMaterializedView(l.ctx, &themeClient.SaveMvReq{
		Name:         req.Name,
		DatasourceId: req.DatasourceId,
		DbName:       req.DbName,
		ViewName:     req.ViewName,
		SelectSql:    req.SelectSql,
		RefreshMode:  req.RefreshMode,
		Beid:         req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return mvToTypes(reply.GetMv()), nil
}
