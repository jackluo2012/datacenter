package logic

import (
	"context"

	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建人群（圈群定义）
func (l *CreateGroupLogic) CreateGroup(in *tag.CreateGroupReq) (*tag.GroupReply, error) {
	// todo: add your logic here and delete this line

	return &tag.GroupReply{}, nil
}
