package logic

import (
	"context"

	"datacenter/common/rpc/common"
	"datacenter/common/rpc/internal/svc"

	"github.com/tal-tech/go-zero/core/logx"
)

type GetBaseAppLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBaseAppLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBaseAppLogic {
	return &GetBaseAppLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetBaseAppLogic) GetBaseApp(in *common.BaseAppReq) (*common.BaseAppResp, error) {
	repl, err := l.svcCtx.BaseAppModel.FindOne(in.Beid)
	if err != nil {
		return nil, err
	}

	return &common.BaseAppResp{
		Beid:        repl.Id,
		Logo:        repl.Logo,
		Fullwebsite: repl.Fullwebsite,
		Isclose:     repl.Isclose,
	}, nil
}
