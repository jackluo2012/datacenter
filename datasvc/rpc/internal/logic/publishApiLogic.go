package logic

import (
	"context"

	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishApiLogic {
	return &PublishApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发布/下线
func (l *PublishApiLogic) PublishApi(in *datasvc.IdReq) (*datasvc.OkReply, error) {
	// todo: add your logic here and delete this line

	return &datasvc.OkReply{}, nil
}
