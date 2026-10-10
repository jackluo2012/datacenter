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

type CreateTagDefLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建标签定义
func NewCreateTagDefLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTagDefLogic {
	return &CreateTagDefLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 创建标签定义: 网关请求逐字段映射为 rpc 请求, 标签类型缺省为 rule(规则标签), 转发后把定义转为网关类型返回
func (l *CreateTagDefLogic) CreateTagDef(req *types.CreateTagDefReq) (resp *types.TagDef, err error) {
	tagType := req.TagType
	if tagType == "" {
		tagType = "rule"
	}
	reply, err := l.svcCtx.TagRpc.CreateTag(l.ctx, &tagClient.CreateTagReq{
		ObjectId:    req.ObjectId,
		TagKey:      req.TagKey,
		TagName:     req.TagName,
		TagType:     tagType,
		RuleSql:     req.RuleSql,
		RefreshCron: req.RefreshCron,
		Beid:        req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return tagDefToTypes(reply.GetDef()), nil
}
