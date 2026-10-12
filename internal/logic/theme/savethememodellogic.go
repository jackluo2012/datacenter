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

type SaveThemeModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存数仓模型（主题下的分层表）
func NewSaveThemeModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveThemeModelLogic {
	return &SaveThemeModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存数仓模型: 网关请求逐字段映射为 rpc 请求, 转发后把模型转为网关类型返回。
// 契约差异: 网关 SaveThemeModelReq 无 Id 字段, rpc SaveThemeModelReq.Id 恒为 0, 即 rpc 侧始终按新建处理;
// create_sql 由 rpc 侧按 fields_json 生成, 网关不透传。
func (l *SaveThemeModelLogic) SaveThemeModel(req *types.SaveThemeModelReq) (resp *types.ThemeModelInfo, err error) {
	reply, err := l.svcCtx.ThemeRpc.SaveThemeModel(l.ctx, &themeClient.SaveThemeModelReq{
		ThemeId:      req.ThemeId,
		Layer:        req.Layer,
		TableName:    req.TableName,
		ModelName:    req.ModelName,
		FieldsJson:   req.FieldsJson,
		DimJson:      req.DimJson,
		DatasourceId: req.DatasourceId,
		DbName:       req.DbName,
		Beid:         req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return themeModelToTypes(reply.GetModel()), nil
}
