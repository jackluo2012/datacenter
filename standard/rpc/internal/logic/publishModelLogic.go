package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishModelLogic {
	return &PublishModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发布（版本+1并落快照）
func (l *PublishModelLogic) PublishModel(in *standard.PublishReq) (*standard.PublishReply, error) {
	// todo: add your logic here and delete this line

	return &standard.PublishReply{}, nil
}
