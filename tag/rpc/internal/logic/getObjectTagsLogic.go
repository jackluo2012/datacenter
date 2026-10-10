package logic

import (
	"context"

	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetObjectTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetObjectTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetObjectTagsLogic {
	return &GetObjectTagsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查对象标签画像
func (l *GetObjectTagsLogic) GetObjectTags(in *tag.ObjectTagsReq) (*tag.ObjectTagsReply, error) {
	// todo: add your logic here and delete this line

	return &tag.ObjectTagsReply{}, nil
}
