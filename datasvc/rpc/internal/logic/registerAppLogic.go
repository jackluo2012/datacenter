package logic

import (
	"context"

	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterAppLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterAppLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterAppLogic {
	return &RegisterAppLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注册应用（返回 app_key/app_secret）
func (l *RegisterAppLogic) RegisterApp(in *datasvc.RegisterAppReq) (*datasvc.AppReply, error) {
	// todo: add your logic here and delete this line

	return &datasvc.AppReply{}, nil
}
