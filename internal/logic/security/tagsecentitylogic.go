// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/security/rpc/securityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type TagSecEntityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 元数据实体分级打标
func NewTagSecEntityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagSecEntityLogic {
	return &TagSecEntityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 实体打标: 网关请求逐字段映射为 rpc 请求(同名字段一一对应), 转发后把打标结果转为网关类型返回。
func (l *TagSecEntityLogic) TagSecEntity(req *types.TagSecEntityReq) (resp *types.SecEntityTag, err error) {
	reply, err := l.svcCtx.SecurityRpc.TagEntity(l.ctx, &securityClient.TagEntityReq{
		EntityId:     req.EntityId,
		EntityName:   req.EntityName,
		LevelCode:    req.LevelCode,
		MaskRuleCode: req.MaskRuleCode,
		Beid:         req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return entityTagToTypes(reply.GetTag()), nil
}
