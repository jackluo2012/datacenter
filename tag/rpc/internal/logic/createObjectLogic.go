package logic

import (
	"context"

	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateObjectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateObjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateObjectLogic {
	return &CreateObjectLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建标签对象
func (l *CreateObjectLogic) CreateObject(in *tag.CreateObjectReq) (*tag.ObjectReply, error) {
	// todo: add your logic here and delete this line

	return &tag.ObjectReply{}, nil
}
