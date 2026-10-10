package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type TagEntityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTagEntityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagEntityLogic {
	return &TagEntityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 实体打标
func (l *TagEntityLogic) TagEntity(in *security.TagEntityReq) (*security.EntityTagReply, error) {
	// todo: add your logic here and delete this line

	return &security.EntityTagReply{}, nil
}
