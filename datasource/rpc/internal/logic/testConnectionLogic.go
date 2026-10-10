package logic

import (
	"context"

	"datacenter/datasource/rpc/datasource"
	"datacenter/datasource/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestConnectionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTestConnectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestConnectionLogic {
	return &TestConnectionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 测试连接
func (l *TestConnectionLogic) TestConnection(in *datasource.IdReq) (*datasource.TestConnReply, error) {
	// todo: add your logic here and delete this line

	return &datasource.TestConnReply{}, nil
}
