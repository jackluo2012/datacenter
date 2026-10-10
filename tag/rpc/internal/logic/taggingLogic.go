package logic

import (
	"context"

	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type TaggingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTaggingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaggingLogic {
	return &TaggingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行打标（规则 SQL 跑目标库，结果写 dc_tag_value）
func (l *TaggingLogic) Tagging(in *tag.TaggingReq) (*tag.TaggingReply, error) {
	// todo: add your logic here and delete this line

	return &tag.TaggingReply{}, nil
}
