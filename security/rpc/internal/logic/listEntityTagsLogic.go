package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListEntityTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListEntityTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEntityTagsLogic {
	return &ListEntityTagsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListEntityTagsLogic) ListEntityTags(in *security.EntityTagListReq) (*security.EntityTagListReply, error) {
	// todo: add your logic here and delete this line

	return &security.EntityTagListReply{}, nil
}
