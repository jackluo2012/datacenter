// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/tag/rpc/tagClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTagDefLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新标签定义
func NewUpdateTagDefLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTagDefLogic {
	return &UpdateTagDefLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateTagDef: 网关请求逐字段映射为 rpc 请求后转发（全量更新，rpc 侧做唯一查重与规则 SQL 校验）,
// status 0.停用/1.启用 为必传契约, 转发后把定义转为网关类型返回
func (l *UpdateTagDefLogic) UpdateTagDef(req *types.UpdateTagDefReq) (resp *types.TagDef, err error) {
	reply, err := l.svcCtx.TagRpc.UpdateTag(l.ctx, &tagClient.UpdateTagReq{
		Id:          req.Id,
		ObjectId:    req.ObjectId,
		TagKey:      req.TagKey,
		TagName:     req.TagName,
		TagType:     req.TagType,
		RuleSql:     req.RuleSql,
		RefreshCron: req.RefreshCron,
		Status:      req.Status,
		Beid:        req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return tagDefToTypes(reply.GetDef()), nil
}
