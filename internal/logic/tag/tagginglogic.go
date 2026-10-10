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

type TaggingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 执行打标
func NewTaggingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaggingLogic {
	return &TaggingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 执行打标: 网关请求逐字段映射为 rpc 请求, 转发后把打标结果转为网关类型返回
func (l *TaggingLogic) Tagging(req *types.TaggingReq) (resp *types.TaggingResp, err error) {
	reply, err := l.svcCtx.TagRpc.Tagging(l.ctx, &tagClient.TaggingReq{
		TagId:    req.TagId,
		ObjectId: req.ObjectId,
		Beid:     req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.TaggingResp{
		TaggedRows: reply.GetTaggedRows(),
		TagId:      reply.GetTagId(),
	}, nil
}
