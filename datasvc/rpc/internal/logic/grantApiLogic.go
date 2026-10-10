package logic

import (
	"context"

	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GrantApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGrantApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrantApiLogic {
	return &GrantApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 授权 API 给应用
func (l *GrantApiLogic) GrantApi(in *datasvc.GrantApiReq) (*datasvc.OkReply, error) {
	// todo: add your logic here and delete this line

	return &datasvc.OkReply{}, nil
}
