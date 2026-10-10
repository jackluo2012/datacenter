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

type ObjectTagsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 对象标签画像
func NewObjectTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ObjectTagsLogic {
	return &ObjectTagsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 对象标签画像: 网关请求逐字段映射为 rpc 请求, 转发后把标签值列表逐项转为网关类型返回
func (l *ObjectTagsLogic) ObjectTags(req *types.ObjectTagsReq) (resp *types.ObjectTagsResp, err error) {
	reply, err := l.svcCtx.TagRpc.GetObjectTags(l.ctx, &tagClient.ObjectTagsReq{
		ObjectPk: req.ObjectPk,
		ObjectId: req.ObjectId,
		Beid:     req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	tags := make([]types.TagValueItem, 0, len(reply.GetTags()))
	for _, item := range reply.GetTags() {
		tags = append(tags, types.TagValueItem{
			TagKey:   item.GetTagKey(),
			TagName:  item.GetTagName(),
			TagValue: item.GetTagValue(),
			TaggedAt: item.GetTaggedAt(),
		})
	}
	return &types.ObjectTagsResp{
		Tags: tags,
	}, nil
}
